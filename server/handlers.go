package server

import (
	"fmt"
	"html"
	"io"
	"mime"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"lanserve/web"
)

type Server struct {
	RootDir    string
	SecretCode string
	Sessions   *SessionStore
}

func NewServer(rootDir, secretCode string) *Server {
	return &Server{
		RootDir:    rootDir,
		SecretCode: secretCode,
		Sessions:   NewSessionStore(),
	}
}

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

	rl.Header().Set("Access-Control-Allow-Origin", "*")
	rl.Header().Set("Access-Control-Allow-Methods", "GET, POST, DELETE, OPTIONS")
	rl.Header().Set("Access-Control-Allow-Headers", "X-Auth-Code, Content-Type")
	rl.Header().Set("Accept-Ranges", "bytes")

	if r.Method == http.MethodOptions {
		rl.WriteHeader(http.StatusNoContent)
		return
	}

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
	filename = strings.Split(filename, "?")[0]

	data, err := web.FS.ReadFile(filepath.Join("static", filename))
	if err != nil {
		http.Error(w, fmt.Sprintf("Static file not found: %s", filename), http.StatusNotFound)
		return
	}

	ext := filepath.Ext(filename)
	contentType := mime.TypeByExtension(ext)
	if ext == ".css" {
		contentType = "text/css; charset=utf-8"
	} else if ext == ".js" {
		contentType = "application/javascript; charset=utf-8"
	}

	if contentType != "" {
		w.Header().Set("Content-Type", contentType)
	}
	w.Header().Set("Cache-Control", "public, max-age=3600")
	w.WriteHeader(http.StatusOK)
	w.Write(data)
}

func (s *Server) handleGet(w http.ResponseWriter, r *http.Request) {
	relPath := strings.TrimPrefix(r.URL.Path, "/")
	decodedPath, err := url.PathUnescape(relPath)
	if err != nil {
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	fullPath := filepath.Join(s.RootDir, decodedPath)
	fi, err := os.Stat(fullPath)
	if err != nil {
		http.Error(w, "Not Found", http.StatusNotFound)
		return
	}

	if fi.IsDir() {
		s.listDirectory(w, r, fullPath, r.URL.Path)
		return
	}

	http.ServeFile(w, r, fullPath)
}

func (s *Server) listDirectory(w http.ResponseWriter, r *http.Request, path, reqPath string) {
	entries, err := os.ReadDir(path)
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

	displayPath, _ := url.PathUnescape(reqPath)

	rootEntries, _ := os.ReadDir(s.RootDir)
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
		link := url.PathEscape(name)
		if isDir {
			link += "/"
		}

		icon := FileIcon(name, isDir)
		sizeStr := ""
		if !isDir {
			sizeStr = HumanSize(info.Size())
		}

		cleanReqPath := strings.TrimRight(reqPath, "/")
		relPath := strings.TrimPrefix(cleanReqPath+"/"+url.PathEscape(name), "/")

		deleteBtn := ""
		if !isDir {
			deleteBtn = fmt.Sprintf(
				`<button class="delete-btn" onclick="event.stopPropagation(); deleteFile('/%s', '%s')">Delete</button>`,
				relPath, html.EscapeString(name),
			)
		}

		displayName := html.EscapeString(name)
		if isDir {
			displayName += "/"
		}

		rows.WriteString(fmt.Sprintf(`
					<tr class="file-row" data-name="%s">
						<td>
							<a href="%s" class="file-link">
								<span class="file-icon">%s</span>
								<span>%s</span>
							</a>
						</td>
						<td style="color: var(--text-muted);">%s</td>
						<td class="text-right">%s</td>
					</tr>`,
			strings.ToLower(displayName), link, icon, displayName, sizeStr, deleteBtn,
		))
	}

	templateBytes, _ := web.FS.ReadFile("template.html")
	htmlContent := string(templateBytes)
	htmlContent = strings.ReplaceAll(htmlContent, "<!-- SLOT:displaypath -->", html.EscapeString(displayPath))
	htmlContent = strings.ReplaceAll(htmlContent, "<!-- SLOT:folder_options -->", folderOpts.String())
	htmlContent = strings.ReplaceAll(htmlContent, "<!-- SLOT:file_items -->", rows.String())

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(htmlContent))
}

func (s *Server) handlePost(w http.ResponseWriter, r *http.Request) {
	authOK, newToken := s.Authenticate(r)
	if !authOK {
		http.Error(w, "Unauthorized access code", http.StatusUnauthorized)
		return
	}
	if newToken != "" {
		http.SetCookie(w, &http.Cookie{
			Name: "lanserve_session", Value: newToken, HttpOnly: true, SameSite: http.SameSiteStrictMode, Path: "/", MaxAge: 31536000,
		})
	}

	if err := r.ParseMultipartForm(128 << 20); err != nil {
		http.Error(w, "Failed to parse form", http.StatusBadRequest)
		return
	}

	targetDir := r.FormValue("target_folder")
	if targetDir == "" {
		targetDir = "."
	}

	if r.MultipartForm != nil && r.MultipartForm.File != nil {
		for _, headers := range r.MultipartForm.File {
			for _, header := range headers {
				filename := filepath.Base(header.Filename)
				uploadPath := filepath.Join(s.RootDir, targetDir, filename)

				os.MkdirAll(filepath.Dir(uploadPath), 0755)
				src, err := header.Open()
				if err != nil {
					continue
				}

				dst, err := os.Create(uploadPath)
				if err != nil {
					src.Close()
					continue
				}

				io.Copy(dst, src)
				src.Close()
				dst.Close()
			}
		}
	}

	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) handleDelete(w http.ResponseWriter, r *http.Request) {
	authOK, newToken := s.Authenticate(r)
	if !authOK {
		http.Error(w, "Unauthorized access code", http.StatusUnauthorized)
		return
	}
	if newToken != "" {
		http.SetCookie(w, &http.Cookie{
			Name: "lanserve_session", Value: newToken, HttpOnly: true, SameSite: http.SameSiteStrictMode, Path: "/", MaxAge: 31536000,
		})
	}

	rel, err := url.PathUnescape(strings.TrimPrefix(r.URL.Path, "/"))
	if err != nil {
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	fullPath := filepath.Clean(filepath.Join(s.RootDir, rel))
	if fullPath != s.RootDir && !IsPathWithin(fullPath, s.RootDir) {
		w.WriteHeader(http.StatusForbidden)
		return
	}

	fi, err := os.Stat(fullPath)
	if err != nil || fi.IsDir() {
		w.WriteHeader(http.StatusForbidden)
		return
	}

	if err := os.Remove(fullPath); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
