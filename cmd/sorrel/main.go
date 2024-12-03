package main

import (
	"fmt"
	"net/url"
	"os"
	"regexp"
	"strings"

	"github.com/atotto/clipboard"
)

func main() {
	// Get the clipboard content
	text, err := clipboard.ReadAll()
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error accessing clipboard:", err)
		return
	}

	fmt.Println("Clipboard content:", text)

	// Validate YouTube URL
	isValid, timestamp, err := validateYouTubeURL(text)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Validation error:", err)
		return
	}

	if isValid {
		fmt.Println("Valid YouTube URL detected.")
		if timestamp != "" {
			fmt.Printf("URL contains timestamp: %s\n", timestamp)
		} else {
			fmt.Println("URL does not contain a timestamp.")
		}
	} else {
		fmt.Println("Invalid YouTube URL.")
	}
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
