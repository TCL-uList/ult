package appdistribution

import (
	"fmt"
	"io"
	"strings"
	"time"
)

var (
	spinner = []string{"⢿", "⡿", "⣟", "⣯", "⣷", "⣾", "⣽", "⣻", "⢿", "⡿", "⣟"}
)

const (
	cleanLine = "\033[2K\r"
)

type ProgressPrinter struct {
	io.Reader
	Total        int64
	Current      int64
	LatestUpdate time.Time
}

// Print the progress of a file upload as %
func (pr *ProgressPrinter) Read(p []byte) (int, error) {
	n, err := pr.Reader.Read(p)
	pr.Current += int64(n)

	now := time.Now()
	if now.Sub(pr.LatestUpdate).Milliseconds() > 300 {
		percent := float64(pr.Current) / float64(pr.Total) * 100
		fmt.Printf("Upload progress: %s\r", fmt.Sprintf("%.1f%%", percent))
		pr.LatestUpdate = now
	}

	return n, err
}

func contentTypeByExtension(extension string) (string, error) {
	PACKAGE := "application/vnd.android.package-archive"
	STREAM := "application/octet-stream"

	switch strings.ToLower(extension) {
	case "apk":
		return PACKAGE, nil
	case "aab":
		return STREAM, nil
	case "ipa":
		return STREAM, nil
	}

	var validExtensions = []string{"apk", "aab", "ipa"}
	return "", fmt.Errorf("the only valid extensions, are: %v, got: %s", validExtensions, extension)
}

// return a new character to print a loading spinner
func printLoadingSpinner() {
	fmt.Printf(cleanLine)
	char := spinner[int(time.Now().UnixNano()/100000000)%len(spinner)]
	fmt.Printf(char)
}

// clean the line that the loading spinner was using
func cleanLoadingSpinner() {
	fmt.Printf(cleanLine)
}
