package main

import (
	"fmt"
	"net/url"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"

	"github.com/atotto/clipboard"
	tea "github.com/charmbracelet/bubbletea"
)

func setupTemporaryLog() (*os.File, error) {
	// Create a temporary file for logging
	tempFile, err := os.CreateTemp("", "sorrel-*.log")
	if err != nil {
		return nil, fmt.Errorf("failed to create temporary log file: %w", err)
	}

	// Redirect log output to the file
	os.Stderr = tempFile
	return tempFile, nil
}

func main() {
	// Set up temporary logging
	logFile, err := setupTemporaryLog()
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error setting up logging:", err)
		os.Exit(1)
	}
	logFile.Close()

	// Open a new kitty instance if not run in an existing one
	if os.Getenv("KITTY_PID") == "" {
		// Not inside Kitty, spawn a new Kitty window
		cmd := exec.Command("kitty", "--class", "sorrel", "sh", "-c", "sorrel")
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			fmt.Fprintln(os.Stderr, "Failed to launch Kitty window:", err)
			os.Exit(1)
		}
		return
	}

	// Get the clipboard content
	text, err := clipboard.ReadAll()
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error accessing clipboard:", err)
		text = "" // Fallback to empty if clipboard fails
	}

	// Validate YouTube URL
	isValid, timestamp, err := validateAndExtract(text)
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

	// Validate the URL
	isValid, err = isValidYouTubeURL(m.urlInput.Value())
	if !isValid || err != nil {
		fmt.Fprintln(os.Stderr, "Validation error:", err)
		return
	}

	// Validate the timestamps
	time1, time2, err := validateTimestamps(m.startTimestamp.Value(), m.endTimestamp.Value())
	if err != nil {
		fmt.Fprintln(os.Stderr, "Validation error:", err)
		return
	}

	// Proceed with further logic using time1 and time2
	fmt.Println("Validation passed:")
	fmt.Printf("URL: %s\nStart Timestamp: %d\nEnd Timestamp: %d\n", m.urlInput.Value(), time1, time2)

}

// isValidYouTubeURL checks if the URL is a valid YouTube link
func isValidYouTubeURL(link string) (bool, error) {
	// Parse the URL
	parsedURL, err := url.Parse(link)
	if err != nil {
		return false, fmt.Errorf("failed to parse URL: %w", err)
	}

	// Check if it’s a YouTube URL
	host := strings.ToLower(parsedURL.Host)
	if host != "www.youtube.com" && host != "youtube.com" && host != "youtu.be" {
		return false, fmt.Errorf("not a YouTube URL")
	}

	return true, nil
}

// extractTimestamp extracts the timestamp from a YouTube URL if present
func extractTimestamp(link string) (string, error) {
	// Parse the URL
	parsedURL, err := url.Parse(link)
	if err != nil {
		return "", fmt.Errorf("failed to parse URL: %w", err)
	}

	// Check for timestamp in query params
	query := parsedURL.Query()
	if ts, exists := query["t"]; exists {
		return ts[0], nil
	}

	// Check for timestamp in URL fragment (e.g., `#t=60s`)
	fragment := parsedURL.Fragment
	re := regexp.MustCompile(`t=(\d+)s?`)
	if matches := re.FindStringSubmatch(fragment); matches != nil {
		return matches[1], nil
	}

	return "", nil // No timestamp found
}

// validateAndExtract checks if the URL is a valid YouTube link and extracts the timestamp if present
func validateAndExtract(link string) (isValid bool, timestamp string, err error) {
	// Validate the URL
	isValid, err = isValidYouTubeURL(link)
	if err != nil || !isValid {
		return false, "", err
	}

	// Extract the timestamp
	timestamp, err = extractTimestamp(link)
	if err != nil {
		return false, "", fmt.Errorf("timestamp extraction failed: %w", err)
	}

	return true, timestamp, nil
}

// Validate first timestamp
func validateStartTimestamp(input string) (int, error) {
	// Parse the input as an integer
	timestamp, err := strconv.Atoi(input)
	if err != nil {
		return 0, fmt.Errorf("invalid start timestamp: %s", input)
	}
	return timestamp, nil
}

// Validate the second timestamp, allow:
// - Integers for absolute timestamps,
// - Negative integers to specify time before the `start` timestamp,
// - `+` prefixed integers to specify time after the `start` timestamp.
func validateEndTimestamp(input string, start int) (int, error) {
	// Check for relative duration (time after `start`)
	if strings.HasPrefix(input, "+") {
		duration, err := strconv.Atoi(input[1:]) // Remove '+' and parse as integer
		if err != nil || duration < 0 {
			return 0, fmt.Errorf("invalid duration format: %s", input)
		}
		return start + duration, nil // Calculate absolute end timestamp
	}

	// Check for negative relative time (time before `start`)
	if strings.HasPrefix(input, "-") {
		duration, err := strconv.Atoi(input) // Parse negative integer directly
		if err != nil || duration >= 0 {     // Ensure it's a valid negative integer
			return 0, fmt.Errorf("invalid negative duration format: %s", input)
		}
		return start + duration, nil // Subtract from start
	}

	// Parse absolute timestamp
	timestamp, err := strconv.Atoi(input)
	if err != nil {
		return 0, fmt.Errorf("invalid end timestamp: %s", input)
	}

	return timestamp, nil // Return absolute timestamp
}

// Validate timestamps, normalize them and reorder if needed
func validateTimestamps(startInput, endInput string) (int, int, error) {
	// Validate the start timestamp
	start, err := validateStartTimestamp(startInput)
	if err != nil {
		return 0, 0, fmt.Errorf("start timestamp validation failed: %w", err)
	}

	// Validate the end timestamp
	end, err := validateEndTimestamp(endInput, start)
	if err != nil {
		return 0, 0, fmt.Errorf("end timestamp validation failed: %w", err)
	}

	// Ensure both timestamps are non-negative
	if start < 0 {
		return 0, 0, fmt.Errorf("start timestamp validation failed: %w", err)
	}
	if end < 0 {
		return 0, 0, fmt.Errorf("end timestamp validation failed: %w", err)
	}

	// Reorder timestamps if needed
	if end < start {
		start, end = end, start
	}

	return start, end, nil
}
