package server

import (
	"encoding/json"
	"errors"
	"fmt"
	"html"
	"io"
	"io/fs"
	"mime"
	"mime/multipart"
	"net/http"
	"os"
	pathpkg "path"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/Dev-syphax/lanserve/web"
)

var (
	errPathEscapes = errors.New("path escapes the shared directory")
	errBadFilename = errors.New("invalid upload filename")
)

type Server struct {
	RootDir    string 
	realRoot   string 
	SecretCode string
	Sessions   *SessionStore
}

func NewServer(rootDir, secretCode string) *Server {
	realRoot := rootDir
	if r, err := filepath.EvalSymlinks(rootDir); err == nil {
		realRoot = r
	}
	return &Server{
		RootDir:    rootDir,
		realRoot:   realRoot,
		SecretCode: secretCode,
		Sessions:   NewSessionStore(),
	}
}


// Path confinement
 

// lexicalJoin joins a slash-separated relative path onto the root and checks
// the cleaned result still lies inside it. It does not touch the filesystem.
func (s *Server) lexicalJoin(rel string) (string, error) {
	if strings.ContainsRune(rel, 0) {
		return "", errPathEscapes
	}
	full := filepath.Join(s.realRoot, filepath.FromSlash(rel))
	if !IsPathWithin(full, s.realRoot) {
		return "", errPathEscapes
	}
	return full, nil
}

// resolveExisting returns the real path of an existing file or directory,
// following symlinks, and refuses anything that ends up outside the root.
func (s *Server) resolveExisting(rel string) (string, error) {
	full, err := s.lexicalJoin(rel)
	if err != nil {
		return "", err
	}
	real, err := filepath.EvalSymlinks(full)
	if err != nil {
		return "", err // does not exist
	}
	if !IsPathWithin(real, s.realRoot) {
		return "", errPathEscapes
	}
	return real, nil
}

// resolveEntry resolves the directory part of rel but not the final element,
// so that deleting a symlink removes the link itself and never its target.
func (s *Server) resolveEntry(rel string) (string, error) {
	full, err := s.lexicalJoin(rel)
	if err != nil {
		return "", err
	}
	dir, base := filepath.Split(full)
	if base == "" || base == "." || base == ".." {
		return "", errPathEscapes
	}
	realDir, err := filepath.EvalSymlinks(dir)
	if err != nil {
		return "", err
	}
	if !IsPathWithin(realDir, s.realRoot) {
		return "", errPathEscapes
	}
	return filepath.Join(realDir, base), nil
}

// resolveDirForWrite returns a real directory path for an upload target. The
// directory may not exist yet: the deepest existing ancestor is resolved
// (symlinks included) and checked before anything is created.
func (s *Server) resolveDirForWrite(rel string) (string, error) {
	full, err := s.lexicalJoin(rel)
	if err != nil {
		return "", err
	}
	existing := full
	for {
		if _, err := os.Lstat(existing); err == nil {
			break
		}
		parent := filepath.Dir(existing)
		if parent == existing {
			return "", errPathEscapes
		}
		existing = parent
	}
	realExisting, err := filepath.EvalSymlinks(existing)
	if err != nil {
		return "", err
	}
	if !IsPathWithin(realExisting, s.realRoot) {
		return "", errPathEscapes
	}
	remainder, err := filepath.Rel(existing, full)
	if err != nil {
		return "", errPathEscapes
	}
	return filepath.Join(realExisting, remainder), nil
}

func writePathError(w http.ResponseWriter, err error) {
	if errors.Is(err, errPathEscapes) {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}
	http.Error(w, "Not Found", http.StatusNotFound)
}

 
// Routing
 

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	start := time.Now()

	// Wrap response writer to track status code
	rl := &responseLogger{
		ResponseWriter: w,
		statusCode:     http.StatusOK, // Default to 200 OK
	}

	// Always execute logger upon function completion
	defer func() {
		LogRequest(r, rl.statusCode, time.Since(start))
	}()

	// No CORS headers on purpose: the web UI is same-origin, and a wildcard
	// policy would let any website the user visits read or modify shared files.
	rl.Header().Set("Accept-Ranges", "bytes")
	rl.Header().Set("X-Content-Type-Options", "nosniff")

	if strings.HasPrefix(r.URL.Path, "/_static/") {
		s.serveStatic(rl, r)
		return
	}

	switch r.Method {
		case http.MethodGet:
			s.handleGet(rl, r)
		case http.MethodPost:
			s.handlePost(rl, r)
		case http.MethodDelete:
			s.handleDelete(rl, r)
		default:
			http.Error(rl, "Method Not Allowed", http.StatusMethodNotAllowed)
	}
}

func (s *Server) serveStatic(w http.ResponseWriter, r *http.Request) {
	filename := strings.TrimPrefix(r.URL.Path, "/_static/")
	if !fs.ValidPath(filename) { // rejects ".."
		http.Error(w, "Not Found", http.StatusNotFound)
		return
	}

	// embed.FS always uses forward slashes, so use "path", not "path/filepath".
	data, err := web.FS.ReadFile(pathpkg.Join("static", filename))
	if err != nil {
		http.Error(w, "Not Found", http.StatusNotFound)
		return
	}

	ext := pathpkg.Ext(filename)
	contentType := mime.TypeByExtension(ext)
	switch ext {
	case ".css":
		contentType = "text/css; charset=utf-8"
	case ".js":
		contentType = "application/javascript; charset=utf-8"
	}

	if contentType != "" {
		w.Header().Set("Content-Type", contentType)
	}
	w.Header().Set("Cache-Control", "public, max-age=3600")
	w.WriteHeader(http.StatusOK)
	w.Write(data)
}

// requireAuth checks the access code / session. On failure it writes the
// response and returns false.
func (s *Server) requireAuth(w http.ResponseWriter, r *http.Request) bool {
	authOK, newToken := s.Authenticate(r)
	if !authOK {
		http.Error(w, "Unauthorized access code", http.StatusUnauthorized)
		return false
	}
	if newToken != "" {
		http.SetCookie(w, &http.Cookie{
			Name: "lanserve_session", Value: newToken, HttpOnly: true, SameSite: http.SameSiteStrictMode, Path: "/", MaxAge: 31536000,
		})
	}
	return true
}

// GET
 
func (s *Server) handleGet(w http.ResponseWriter, r *http.Request) {
	// r.URL.Path has already been percent-decoded exactly once by net/http.
	// Decoding it again is what made "%252e%252e" behave like "..".
	rel := strings.TrimPrefix(r.URL.Path, "/")

	fullPath, err := s.resolveExisting(rel)
	if err != nil {
		writePathError(w, err)
		return
	}

	fi, err := os.Stat(fullPath)
	if err != nil {
		http.Error(w, "Not Found", http.StatusNotFound)
		return
	}

	if fi.IsDir() {
		urlRel := strings.Trim(pathpkg.Clean("/"+r.URL.Path), "/")
		s.listDirectory(w, fullPath, urlRel)
		return
	}

	http.ServeFile(w, r, fullPath)
}

// jsString renders s as a JavaScript string literal that is safe inside a
// double-quoted HTML attribute: JSON escapes quotes/backslashes/control
// characters for the JS parser, then HTML-escaping protects the attribute.
func jsString(s string) string {
	b, _ := json.Marshal(s)
	return html.EscapeString(string(b))
}

func hasImageExt(lowerName string) bool {
	for _, ext := range []string{".png", ".jpg", ".jpeg", ".gif", ".bmp", ".webp"} {
		if strings.HasSuffix(lowerName, ext) {
			return true
		}
	}
	return false
}

// listDirectory renders dir. rel is the cleaned, decoded, slash-separated
// location below the root ("" for the root); nothing taken from the raw
// request is placed in the page.
func (s *Server) listDirectory(w http.ResponseWriter, dir, rel string) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		http.Error(w, "Cannot list directory", http.StatusNotFound)
		return
	}

	sort.Slice(entries, func(i, j int) bool {
		if entries[i].IsDir() != entries[j].IsDir() {
			return entries[i].IsDir()
		}
		return strings.ToLower(entries[i].Name()) < strings.ToLower(entries[j].Name())
	})

	displayPath := "/"
	if rel != "" {
		displayPath = "/" + rel + "/"
	}

	rootEntries, _ := os.ReadDir(s.realRoot)
	var folderOpts strings.Builder
	folderOpts.WriteString(`<option value=".">Root Directory (/)</option>`)
	for _, entry := range rootEntries {
		if entry.IsDir() && !strings.HasPrefix(entry.Name(), ".") {
			name := html.EscapeString(entry.Name())
			folderOpts.WriteString(fmt.Sprintf(`<option value="%s">%s/</option>`, name, name))
		}
	}

	var rows strings.Builder
	for _, entry := range entries {
		name := entry.Name()
		info, err := entry.Info()
		if err != nil {
			continue
		}

		isDir := entry.IsDir()
		isImg := hasImageExt(strings.ToLower(name))

		entryRel := pathpkg.Join(rel, name)
		entryURL := escapedURLPath(entryRel) // absolute and escaped: "/a%20b/c.txt"
		link := entryURL
		if isDir {
			link += "/"
		}

		sizeStr := ""
		if !isDir {
			sizeStr = HumanSize(info.Size())
		}

		deleteBtn := ""
		if !isDir {
			deleteBtn = fmt.Sprintf(
				`<button class="delete-btn" onclick="event.stopPropagation(); deleteFile(%s, %s)">Delete</button>`,
						jsString(entryURL), jsString(name),
			)
		}

		displayName := name
		if isDir {
			displayName += "/"
		}

		var thumb string
		if isImg {
			thumb = fmt.Sprintf(`<img src="%s" alt="%s" class="file-thumbnail">`, html.EscapeString(link), html.EscapeString(name))
		} else {
			thumb = fmt.Sprintf(`<span class="file-icon">%s</span>`, FileIcon(name, isDir))
		}

		rows.WriteString(fmt.Sprintf(`
		<tr class="file-row" data-name="%s">
		<td>
		<a href="%s" class="file-link">
		<div>%s</div>
		<span>%s</span>
		</a>
		</td>
		<td style="color: var(--text-muted);">%s</td>
		<td class="text-right">%s</td>
		</tr>`,
		html.EscapeString(strings.ToLower(displayName)),
					     html.EscapeString(link),
					     thumb,
			       html.EscapeString(displayName),
					     sizeStr,
			       deleteBtn,
		))
	}

	templateBytes, err := web.FS.ReadFile("template.html")
	if err != nil {
		http.Error(w, "Template missing", http.StatusInternalServerError)
		return
	}
	htmlContent := string(templateBytes)
	htmlContent = strings.ReplaceAll(htmlContent, "<!-- SLOT:displaypath -->", html.EscapeString(displayPath))
	htmlContent = strings.ReplaceAll(htmlContent, "<!-- SLOT:folder_options -->", folderOpts.String())
	htmlContent = strings.ReplaceAll(htmlContent, "<!-- SLOT:file_items -->", rows.String())

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(htmlContent))
}

 
// POST (upload)
 

func saveUpload(dir string, header *multipart.FileHeader) error {
	// Normalise Windows separators first so "..\\..\\x" is reduced to "x" on
	// every platform, then keep only the final element.
	name := pathpkg.Base(strings.ReplaceAll(header.Filename, "\\", "/"))
	if name == "." || name == ".." || name == "/" || strings.ContainsRune(name, 0) {
		return errBadFilename
	}
	dest := filepath.Join(dir, name)

	// Never write through a symlink (or into a directory / device) that is
	// already sitting at the destination.
	if fi, err := os.Lstat(dest); err == nil && !fi.Mode().IsRegular() {
		return errors.New("refusing to overwrite a non-regular file")
	}

	src, err := header.Open()
	if err != nil {
		return err
	}
	defer src.Close()

	dst, err := os.OpenFile(dest, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0o644)
	if err != nil {
		return err
	}
	if _, err := io.Copy(dst, src); err != nil {
		dst.Close()
		return err
	}
	return dst.Close()
}

func (s *Server) handlePost(w http.ResponseWriter, r *http.Request) {
	if !s.requireAuth(w, r) {
		return
	}

	if err := r.ParseMultipartForm(128 << 20); err != nil {
		http.Error(w, "Failed to parse form", http.StatusBadRequest)
		return
	}

	targetRel := r.FormValue("target_folder")
	if targetRel == "" {
		targetRel = "."
	}

	// Validate the destination BEFORE creating anything on disk.
	targetDir, err := s.resolveDirForWrite(targetRel)
	if err != nil {
		writePathError(w, err)
		return
	}
	if err := os.MkdirAll(targetDir, 0o755); err != nil {
		http.Error(w, "Cannot create target folder", http.StatusInternalServerError)
		return
	}

	failed := false
	if r.MultipartForm != nil && r.MultipartForm.File != nil {
		for _, headers := range r.MultipartForm.File {
			for _, header := range headers {
				if err := saveUpload(targetDir, header); err != nil {
					failed = true
				}
			}
		}
	}

	if failed {
		http.Error(w, "One or more files could not be saved", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// DELETE

func (s *Server) handleDelete(w http.ResponseWriter, r *http.Request) {
	if !s.requireAuth(w, r) {
		return
	}

	rel := strings.TrimPrefix(r.URL.Path, "/")
	target, err := s.resolveEntry(rel)
	if err != nil {
		writePathError(w, err)
		return
	}

	fi, err := os.Lstat(target)
	if err != nil {
		http.Error(w, "Not Found", http.StatusNotFound)
		return
	}
	if fi.IsDir() {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	if err := os.Remove(target); err != nil {
		http.Error(w, "Delete failed", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
