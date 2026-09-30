package main

import (
	"flag"
	"fmt"
	"github.com/Dev-syphax/lanserve/server"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"unicode/utf8"
)

// Regex to strip ANSI escape codes when calculating visual width
var ansiRegex = regexp.MustCompile(`\x1b\[[0-9;]*m`)

// visibleLen calculates exact rendered character count in terminal
func visibleLen(s string) int {
	clean := ansiRegex.ReplaceAllString(s, "")
	return utf8.RuneCountInString(clean)
}

// truncatePath truncates long paths in the middle so they fit nicely
func truncatePath(path string, maxLen int) string {
	if utf8.RuneCountInString(path) <= maxLen {
		return path
	}
	if maxLen <= 5 {
		return path[:maxLen]
	}
	runes := []rune(path)
	half := (maxLen - 3) / 2
	return string(runes[:half]) + "..." + string(runes[len(runes)-half:])
}

// printBoxRow prints a row with precise dynamic padding
func printBoxRow(label, value string, boxWidth int) {
	// 3 spaces left margin, 3 spaces right margin inside border
	content := fmt.Sprintf("│   %s%s", label, value)
	visWidth := visibleLen(content)

	// Calculate padding needed to reach closing border
	padding := boxWidth - visWidth - 1 // -1 for closing border '│'
	if padding < 3 {
		padding = 3
	}

	fmt.Printf("\033[36m%s\033[0m%s\033[36m│\033[0m\n", content, strings.Repeat(" ", padding))
}

func printBanner(host string, port int, ip, dir, code string) {
	const boxWidth = 60 // Fixed total width of the box

	authStatus := "\033[31mDisabled\033[0m"
	if code != "" {
		authStatus = "\033[32mEnabled\033[0m"
	}

	// Maximum allowed visible length for path before truncating
	// Total width - margins(6) - label("Serving: "(9)) - border(1)
	maxPathLen := boxWidth - 16
	safeDir := truncatePath(dir, maxPathLen)

	topBorder := fmt.Sprintf("┌%s┐", strings.Repeat("─", boxWidth-2))
	midBorder := fmt.Sprintf("├%s┤", strings.Repeat("─", boxWidth-2))
	botBorder := fmt.Sprintf("└%s┘", strings.Repeat("─", boxWidth-2))

	fmt.Printf("\033[36m%s\033[0m\n", topBorder)
	printBoxRow("\033[1;35m LANserve\033[0m — Local File Server", "", boxWidth)
	fmt.Printf("\033[36m%s\033[0m\n", midBorder)

	printBoxRow("\033[1mLocal:\033[0m   ", fmt.Sprintf("http://localhost:%d", port), boxWidth)
	printBoxRow("\033[1mNetwork:\033[0m ", fmt.Sprintf("http://%s:%d", ip, port), boxWidth)
	printBoxRow("\033[1mServing:\033[0m ", safeDir, boxWidth)
	printBoxRow("\033[1mAuth:\033[0m    ", authStatus, boxWidth)

	fmt.Printf("\033[36m%s\033[0m\n", botBorder)
	fmt.Println("\033[90m[TIME]    │ STATUS │ LATENCY │ CLIENT IP       │ REQUEST\033[0m")
	fmt.Println("\033[90m──────────┼────────┼─────────┼─────────────────┼────────────────────────\033[0m")
}

func main() {
	port := flag.Int("port", 8080, "Port to listen on")
	flag.IntVar(port, "p", 8080, "Port short alias")

	host := flag.String("host", "0.0.0.0", "Address to bind to")

	dir := flag.String("dir", ".", "Directory to serve")
	flag.StringVar(dir, "d", ".", "Dir short alias")

	code := flag.String("code", "", "Access code required for uploads/deletes")
	flag.StringVar(code, "c", "", "Code short alias")

	flag.Parse()

	absDir, err := filepath.Abs(*dir)
	if err != nil {
		fmt.Printf("Error resolving directory path: %v\n", err)
		os.Exit(1)
	}

	ip := server.GetLocalIP()
	srv := server.NewServer(absDir, *code)

	printBanner(*host, *port, ip, absDir, *code)

	addr := fmt.Sprintf("%s:%d", *host, *port)
	httpServer := &http.Server{
		Addr:    addr,
		Handler: srv,
	}

	if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		fmt.Printf("Server failed: %v\n", err)
	}
}
