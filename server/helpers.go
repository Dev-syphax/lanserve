package server

import (
	"fmt"
	"net"
	"path/filepath"
	"strings"
)

func GetLocalIP() string {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		return "localhost"
	}
	defer conn.Close()
	return conn.LocalAddr().(*net.UDPAddr).IP.String()
}

func HumanSize(size int64) string {
	sf := float64(size)
	units := []string{"B", "KB", "MB", "GB"}
	for _, unit := range units {
		if sf < 1024 {
			return fmt.Sprintf("%.1f %s", sf, unit)
		}
		sf /= 1024
	}
	return fmt.Sprintf("%.1f TB", sf)
}

func FileIcon(name string, isDir bool) string {
	if isDir {
		return "📁"
	}
	ext := strings.ToLower(filepath.Ext(name))
	if len(ext) > 0 {
		ext = ext[1:]
	}
	icons := map[string]string{
		"py": "🐍", "js": "📜", "ts": "📜", "jsx": "📜", "tsx": "📜",
		"json": "📋", "md": "📝", "txt": "📝", "html": "🌐", "css": "🎨",
		"png": "🖼️", "jpg": "🖼️", "jpeg": "🖼️", "gif": "🖼️", "svg": "🖼️", "webp": "🖼️",
		"mp4": "🎬", "mov": "🎬", "mkv": "🎬", "mp3": "🎵", "wav": "🎵",
		"zip": "📦", "tar": "📦", "gz": "📦", "pdf": "📄", "go": "🐹",
		"sh": "⚙️", "env": "🔑",
	}
	if icon, ok := icons[ext]; ok {
		return icon
	}
	return "📄"
}

func IsPathWithin(target, base string) bool {
	rel, err := filepath.Rel(base, target)
	if err != nil {
		return false
	}
	return !strings.HasPrefix(rel, "..") && rel != ".." && rel != "."
}
