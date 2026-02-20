package wsl

import (
	"encoding/json"
	"fmt"
	"net/http"
	"runtime"
	"strings"

	"github.com/minio/selfupdate"
)

type GitHubRelease struct {
	TagName string `json:"tag_name"`
	Assets  []struct {
		Name               string `json:"name"`
		BrowserDownloadURL string `json:"browser_download_url"`
	} `json:"assets"`
}

// SelfUpdate checks for a new version on GitHub and applies it
func SelfUpdate(currentVersion string) error {
	fmt.Printf("Checking for updates (current version: %s)...\n", currentVersion)

	resp, err := http.Get("https://api.github.com/repos/wendelmax/ezship/releases/latest")
	if err != nil {
		return fmt.Errorf("failed to check for updates: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("github api returned status: %s", resp.Status)
	}

	var release GitHubRelease
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return fmt.Errorf("failed to parse github response: %w", err)
	}

	if release.TagName == currentVersion {
		fmt.Println("You are already using the latest version.")
		return nil
	}

	fmt.Printf("New version available: %s. Downloading...\n", release.TagName)

	// Determine asset name (e.g., ezship-amd64.exe)
	arch := runtime.GOARCH
	if arch == "amd64" {
		// handle possible variations if needed
	}
	targetAsset := fmt.Sprintf("ezship-%s.exe", arch)

	var downloadURL string
	for _, asset := range release.Assets {
		if strings.EqualFold(asset.Name, targetAsset) {
			downloadURL = asset.BrowserDownloadURL
			break
		}
	}

	if downloadURL == "" {
		return fmt.Errorf("could not find binary for your architecture: %s", targetAsset)
	}

	// Download and apply update
	resp, err = http.Get(downloadURL)
	if err != nil {
		return fmt.Errorf("failed to download update: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to download: github returned status %s", resp.Status)
	}

	err = selfupdate.Apply(resp.Body, selfupdate.Options{})
	if err != nil {
		return fmt.Errorf("failed to apply update: %w", err)
	}

	fmt.Printf("Successfully updated to %s! Please restart ezship.\n", release.TagName)
	return nil
}
