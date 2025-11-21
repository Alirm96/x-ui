package service

import (
	"archive/tar"
	"archive/zip"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"time"

	"github.com/alireza0/x-ui/logger"
)

type XrayUpdateService struct{}

// GitHubRelease represents a GitHub release
type GitHubRelease struct {
	TagName    string `json:"tag_name"`
	Name       string `json:"name"`
	Body       string `json:"body"`
	Prerelease bool   `json:"prerelease"`
	Assets     []struct {
		Name               string `json:"name"`
		BrowserDownloadURL string `json:"browser_download_url"`
		Size               int64  `json:"size"`
	} `json:"assets"`
	PublishedAt time.Time `json:"published_at"`
}

// VersionInfo represents version information
type VersionInfo struct {
	Current     string    `json:"current"`
	Latest      string    `json:"latest"`
	HasUpdate   bool      `json:"hasUpdate"`
	ReleaseURL  string    `json:"releaseUrl"`
	DownloadURL string    `json:"downloadUrl"`
	ReleaseDate time.Time `json:"releaseDate"`
	Size        int64     `json:"size"`
	Changelog   string    `json:"changelog"`
}

const (
	xrayGitHubAPI = "https://api.github.com/repos/XTLS/Xray-core/releases"
	xrayTempPath  = "/tmp/xray-update"
)

// getXrayBinPath returns the path to the xray binary based on architecture
func getXrayBinPath() string {
	archMap := map[string]string{
		"amd64": "amd64",
		"386":   "i386",
		"arm64": "arm64",
		"arm":   "arm",
	}
	
	arch, ok := archMap[runtime.GOARCH]
	if !ok {
		arch = "amd64"
	}
	
	return fmt.Sprintf("/app/bin/xray-linux-%s", arch)
}

// GetCurrentVersion gets the currently installed Xray version
func (s *XrayUpdateService) GetCurrentVersion() (string, error) {
	xrayBinPath := getXrayBinPath()
	
	// Try to execute xray -version
	cmd := exec.Command(xrayBinPath, "-version")
	output, err := cmd.CombinedOutput()
	if err != nil {
		logger.Warning("Failed to get xray version:", err)
		return "", fmt.Errorf("failed to execute xray: %v", err)
	}

	// Parse version from output
	// Expected format: "Xray 1.8.7 (Xray, Penetrates Everything.) Custom (go1.21.5 linux/amd64)"
	re := regexp.MustCompile(`Xray\s+(\d+\.\d+\.\d+)`)
	matches := re.FindStringSubmatch(string(output))
	if len(matches) < 2 {
		return "", fmt.Errorf("failed to parse version from output: %s", string(output))
	}

	return matches[1], nil
}

// CheckForUpdates checks if there's a new version available
func (s *XrayUpdateService) CheckForUpdates(includePrerelease bool) (*VersionInfo, error) {
	currentVersion, err := s.GetCurrentVersion()
	if err != nil {
		logger.Warning("Could not determine current version:", err)
		currentVersion = "unknown"
	}

	// Fetch latest release from GitHub
	url := xrayGitHubAPI
	if includePrerelease {
		// Get all releases including prereleases
		url = xrayGitHubAPI
	} else {
		// Get latest stable release only
		url = xrayGitHubAPI + "/latest"
	}

	client := &http.Client{Timeout: 30 * time.Second}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %v", err)
	}

	req.Header.Set("Accept", "application/vnd.github.v3+json")
	req.Header.Set("User-Agent", "x-ui")

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch releases: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GitHub API returned status %d", resp.StatusCode)
	}

	var release GitHubRelease
	var releases []GitHubRelease

	if includePrerelease {
		// Parse array of releases
		if err := json.NewDecoder(resp.Body).Decode(&releases); err != nil {
			return nil, fmt.Errorf("failed to parse releases: %v", err)
		}
		if len(releases) == 0 {
			return nil, fmt.Errorf("no releases found")
		}
		release = releases[0]
	} else {
		// Parse single latest release
		if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
			return nil, fmt.Errorf("failed to parse release: %v", err)
		}
	}

	// Get download URL for current platform
	downloadURL, size, err := s.getDownloadURL(&release)
	if err != nil {
		return nil, err
	}

	// Compare versions
	latestVersion := strings.TrimPrefix(release.TagName, "v")
	hasUpdate := s.compareVersions(currentVersion, latestVersion)

	versionInfo := &VersionInfo{
		Current:     currentVersion,
		Latest:      latestVersion,
		HasUpdate:   hasUpdate,
		ReleaseURL:  fmt.Sprintf("https://github.com/XTLS/Xray-core/releases/tag/%s", release.TagName),
		DownloadURL: downloadURL,
		ReleaseDate: release.PublishedAt,
		Size:        size,
		Changelog:   release.Body,
	}

	return versionInfo, nil
}

// getDownloadURL finds the appropriate download URL for the current platform
func (s *XrayUpdateService) getDownloadURL(release *GitHubRelease) (string, int64, error) {
	// Determine platform and architecture
	osName := runtime.GOOS
	archName := runtime.GOARCH

	// Map Go arch names to Xray release names
	archMap := map[string]string{
		"amd64": "64",
		"386":   "32",
		"arm64": "arm64-v8a",
		"arm":   "arm32-v7a",
	}

	xrayArch, ok := archMap[archName]
	if !ok {
		return "", 0, fmt.Errorf("unsupported architecture: %s", archName)
	}

	// Build expected filename pattern
	// Example: Xray-linux-64.zip
	pattern := fmt.Sprintf("Xray-%s-%s.zip", osName, xrayArch)

	// Find matching asset
	for _, asset := range release.Assets {
		if asset.Name == pattern {
			return asset.BrowserDownloadURL, asset.Size, nil
		}
	}

	return "", 0, fmt.Errorf("no suitable release asset found for %s/%s (looking for %s)", osName, archName, pattern)
}

// compareVersions returns true if latestVersion is greater than currentVersion
func (s *XrayUpdateService) compareVersions(current, latest string) bool {
	if current == "unknown" {
		return true
	}

	// Simple semantic version comparison
	currentParts := strings.Split(current, ".")
	latestParts := strings.Split(latest, ".")

	for i := 0; i < len(currentParts) && i < len(latestParts); i++ {
		var currentNum, latestNum int
		fmt.Sscanf(currentParts[i], "%d", &currentNum)
		fmt.Sscanf(latestParts[i], "%d", &latestNum)

		if latestNum > currentNum {
			return true
		} else if latestNum < currentNum {
			return false
		}
	}

	return len(latestParts) > len(currentParts)
}

// DownloadAndInstall downloads and installs the new Xray version
func (s *XrayUpdateService) DownloadAndInstall(downloadURL string, progressCallback func(downloaded, total int64)) error {
	// Create temp directory
	if err := os.MkdirAll(xrayTempPath, 0755); err != nil {
		return fmt.Errorf("failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(xrayTempPath)

	// Download file
	zipPath := filepath.Join(xrayTempPath, "xray.zip")
	if err := s.downloadFile(downloadURL, zipPath, progressCallback); err != nil {
		return fmt.Errorf("failed to download: %v", err)
	}

	// Extract archive
	extractPath := filepath.Join(xrayTempPath, "extracted")
	if err := os.MkdirAll(extractPath, 0755); err != nil {
		return fmt.Errorf("failed to create extract directory: %v", err)
	}

	if err := s.extractZip(zipPath, extractPath); err != nil {
		return fmt.Errorf("failed to extract: %v", err)
	}

	// Find xray binary in extracted files
	xrayBinary := filepath.Join(extractPath, "xray")
	if _, err := os.Stat(xrayBinary); os.IsNotExist(err) {
		return fmt.Errorf("xray binary not found in archive")
	}

	// Backup current xray
	xrayBinPath := getXrayBinPath()
	backupPath := xrayBinPath + ".backup"
	if _, err := os.Stat(xrayBinPath); err == nil {
		if err := os.Rename(xrayBinPath, backupPath); err != nil {
			return fmt.Errorf("failed to backup current xray: %v", err)
		}
		logger.Info("Backed up current xray to:", backupPath)
	}

	// Install new xray
	if err := s.copyFile(xrayBinary, xrayBinPath); err != nil {
		// Restore backup on failure
		if _, err := os.Stat(backupPath); err == nil {
			os.Rename(backupPath, xrayBinPath)
		}
		return fmt.Errorf("failed to install new xray: %v", err)
	}

	// Set executable permissions
	if err := os.Chmod(xrayBinPath, 0755); err != nil {
		return fmt.Errorf("failed to set permissions: %v", err)
	}

	logger.Info("Successfully installed new Xray version")

	// Remove backup after successful installation
	os.Remove(backupPath)

	return nil
}

// downloadFile downloads a file with progress callback
func (s *XrayUpdateService) downloadFile(url, filepath string, progressCallback func(downloaded, total int64)) error {
	client := &http.Client{Timeout: 10 * time.Minute}
	resp, err := client.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("bad status: %s", resp.Status)
	}

	out, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer out.Close()

	// Progress tracking
	totalSize := resp.ContentLength
	downloaded := int64(0)

	buffer := make([]byte, 32*1024)
	for {
		n, err := resp.Body.Read(buffer)
		if n > 0 {
			out.Write(buffer[:n])
			downloaded += int64(n)
			if progressCallback != nil {
				progressCallback(downloaded, totalSize)
			}
		}
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
	}

	return nil
}

// extractZip extracts a zip file to a destination directory
func (s *XrayUpdateService) extractZip(zipPath, destPath string) error {
	reader, err := zip.OpenReader(zipPath)
	if err != nil {
		return err
	}
	defer reader.Close()

	for _, file := range reader.File {
		filePath := filepath.Join(destPath, file.Name)

		if file.FileInfo().IsDir() {
			os.MkdirAll(filePath, file.Mode())
			continue
		}

		if err := os.MkdirAll(filepath.Dir(filePath), 0755); err != nil {
			return err
		}

		outFile, err := os.OpenFile(filePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, file.Mode())
		if err != nil {
			return err
		}

		rc, err := file.Open()
		if err != nil {
			outFile.Close()
			return err
		}

		_, err = io.Copy(outFile, rc)
		outFile.Close()
		rc.Close()

		if err != nil {
			return err
		}
	}

	return nil
}

// extractTarGz extracts a tar.gz file to a destination directory
func (s *XrayUpdateService) extractTarGz(tarGzPath, destPath string) error {
	file, err := os.Open(tarGzPath)
	if err != nil {
		return err
	}
	defer file.Close()

	gzReader, err := gzip.NewReader(file)
	if err != nil {
		return err
	}
	defer gzReader.Close()

	tarReader := tar.NewReader(gzReader)

	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		filePath := filepath.Join(destPath, header.Name)

		switch header.Typeflag {
		case tar.TypeDir:
			if err := os.MkdirAll(filePath, 0755); err != nil {
				return err
			}
		case tar.TypeReg:
			if err := os.MkdirAll(filepath.Dir(filePath), 0755); err != nil {
				return err
			}

			outFile, err := os.Create(filePath)
			if err != nil {
				return err
			}

			if _, err := io.Copy(outFile, tarReader); err != nil {
				outFile.Close()
				return err
			}
			outFile.Close()

			if err := os.Chmod(filePath, os.FileMode(header.Mode)); err != nil {
				return err
			}
		}
	}

	return nil
}

// copyFile copies a file from src to dst
func (s *XrayUpdateService) copyFile(src, dst string) error {
	sourceFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer sourceFile.Close()

	destFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destFile.Close()

	_, err = io.Copy(destFile, sourceFile)
	return err
}

// RestartXray restarts the Xray service after update
func (s *XrayUpdateService) RestartXray() error {
	// This will be handled by the existing restart mechanism
	xrayService := &XrayService{}
	return xrayService.RestartXray(true)
}

// GetUpdateHistory returns the changelog/release notes
func (s *XrayUpdateService) GetUpdateHistory(version string) (string, error) {
	url := fmt.Sprintf("%s/tags/v%s", xrayGitHubAPI, version)

	client := &http.Client{Timeout: 30 * time.Second}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", err
	}

	req.Header.Set("Accept", "application/vnd.github.v3+json")
	req.Header.Set("User-Agent", "x-ui")

	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("GitHub API returned status %d", resp.StatusCode)
	}

	var release GitHubRelease
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return "", err
	}

	return release.Body, nil
}
