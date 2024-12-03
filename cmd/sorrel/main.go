package main

import (
	"fmt"
	"net/url"
	"os"
	"regexp"
	"strings"

	"github.com/atotto/clipboard"
	tea "github.com/charmbracelet/bubbletea"
)

func main() {
	// Get the clipboard content
	text, err := clipboard.ReadAll()
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error accessing clipboard:", err)
		text = "" // Fallback to empty if clipboard fails
	}

	fmt.Println("Clipboard content:", text)

	// Validate YouTube URL
	isValid, timestamp, err := validateYouTubeURL(text)
	if err != nil || !isValid {
		fmt.Fprintln(os.Stderr, "Validation error:", err)
		text = ""      // Clear the text for invalid or failed validation
		timestamp = "" // Clear the timestamp as well
	}

	// Start the Bubble Tea UI
	p := tea.NewProgram(initialModel(text, timestamp))
	finalModel, err := p.Run()
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error running UI:", err)
		os.Exit(1)
	}

	m := finalModel.(model)
	fmt.Println("Final URL:", m.urlInput.Value())
	fmt.Println("Start Timestamp:", m.startTimestamp.Value())
	fmt.Println("End Timestamp:", m.endTimestamp.Value())
}

// validateYouTubeURL checks if the URL is a valid YouTube link and extracts timestamp if present
func validateYouTubeURL(link string) (isValid bool, timestamp string, err error) {
	// Parse the URL
	parsedURL, err := url.Parse(link)
	if err != nil {
		return false, "", fmt.Errorf("failed to parse URL: %w", err)
	}

	// Check if it’s a YouTube URL
	host := strings.ToLower(parsedURL.Host)
	if host != "www.youtube.com" && host != "youtube.com" && host != "youtu.be" {
		return false, "", fmt.Errorf("not a YouTube URL")
	}

	// Check for timestamp in query params
	query := parsedURL.Query()
	if ts, exists := query["t"]; exists {
		return true, ts[0], nil
	}

	// Check for timestamp in URL fragment (e.g., `#t=60s`)
	fragment := parsedURL.Fragment
	re := regexp.MustCompile(`t=(\d+)s?`)
	if matches := re.FindStringSubmatch(fragment); matches != nil {
		return true, matches[1], nil
	}

	return true, "", nil
}
