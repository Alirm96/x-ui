package util

import (
	"fmt"
	"os"
	"runtime"
	"net/http"
	"io"
)

// DownloadFile downloads a file from the given URL to the given filepath
func DownloadFile(filepath string, url string) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return fmt.Errorf("failed to download %s: status %d", url, resp.StatusCode)
	}

	out, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	return err
}

// GetArch returns the current architecture string for xray download
func GetArch() string {
	switch runtime.GOARCH {
	case "amd64":
		return "64"
	case "arm64":
		return "arm64-v8a"
	case "arm":
		return "arm32-v7a"
	case "386":
		return "32"
	default:
		return "64"
	}
}
