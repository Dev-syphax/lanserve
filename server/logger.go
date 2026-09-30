package server

import (
	"fmt"
	"net"
	"net/http"
	"strings"
	"time"
)

// ANSI Color Codes
const (
	colorReset   = "\033[0m"
	colorBold    = "\033[1m"
	colorDim     = "\033[2m"
	colorRed     = "\033[31m"
	colorGreen   = "\033[32m"
	colorYellow  = "\033[33m"
	colorBlue    = "\033[34m"
	colorMagenta = "\033[35m"
	colorCyan    = "\033[36m"
	colorGray    = "\033[90m"
)

// responseLogger intercepts WriteHeader to capture the HTTP status code
type responseLogger struct {
	http.ResponseWriter
	statusCode int
}

func (rl *responseLogger) WriteHeader(code int) {
	rl.statusCode = code
	rl.ResponseWriter.WriteHeader(code)
}

// Extract clean client IP (handles proxies like X-Forwarded-For if present)
func getClientIP(r *http.Request) string {
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		parts := strings.Split(xff, ",")
		return strings.TrimSpace(parts[0])
	}
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	if host == "::1" || host == "127.0.0.1" {
		return "localhost"
	}
	return host
}

// Colorize HTTP Method badges
func colorizeMethod(method string) string {
	switch method {
	case http.MethodGet:
		return fmt.Sprintf("%s%s %-6s%s", colorBlue, colorBold, method, colorReset)
	case http.MethodPost:
		return fmt.Sprintf("%s%s %-6s%s", colorGreen, colorBold, method, colorReset)
	case http.MethodDelete:
		return fmt.Sprintf("%s%s %-6s%s", colorRed, colorBold, method, colorReset)
	case http.MethodOptions:
		return fmt.Sprintf("%s%s %-6s%s", colorGray, colorBold, method, colorReset)
	default:
		return fmt.Sprintf("%s%s %-6s%s", colorMagenta, colorBold, method, colorReset)
	}
}

// Colorize HTTP Status Codes
func colorizeStatus(status int) string {
	switch {
	case status >= 200 && status < 300:
		return fmt.Sprintf("%s%s %d %s", colorGreen, colorBold, status, colorGray)
	case status >= 300 && status < 400:
		return fmt.Sprintf("%s%s %d %s", colorCyan, colorBold, status, colorGray)
	case status >= 400 && status < 500:
		return fmt.Sprintf("%s%s %d %s", colorYellow, colorBold, status, colorGray)
	default:
		return fmt.Sprintf("%s%s %d %s", colorRed, colorBold, status, colorGray)
	}
}

// LogRequest outputs formatted CLI request logs
func LogRequest(r *http.Request, status int, duration time.Duration) {
	timestamp := time.Now().Format("15:04:05")
	ip := getClientIP(r)
	method := colorizeMethod(r.Method)
	statusCode := colorizeStatus(status)

	// Format latency
	var latency string
	if duration < time.Millisecond {
		latency = fmt.Sprintf("%6.2fµs", float64(duration.Microseconds()))
	} else if duration < time.Second {
		latency = fmt.Sprintf("%6.2fms", float64(duration.Milliseconds()))
	} else {
		latency = fmt.Sprintf("%6.2fs ", duration.Seconds())
	}

	fmt.Printf(" %s%s%s │ %s  │%s%s │ %s%-15s%s │ %s %s\n",
		colorGray, timestamp, colorGray,
		statusCode,
		latency, colorGray,
		colorCyan, ip, colorGray,
		method,
		r.URL.Path,
	)
}
