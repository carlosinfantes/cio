// Package remote implements plugin downloading functionality.
package remote

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

// Downloader handles downloading and extracting plugins.
type Downloader struct {
	client *Client
}

// NewDownloader creates a new plugin downloader.
func NewDownloader(client *Client) *Downloader {
	return &Downloader{client: client}
}

// Download downloads and extracts a plugin to the specified directory.
func (d *Downloader) Download(domain string, destDir string) error {
	// Get plugin info from registry
	plugin, err := d.client.GetPlugin(domain)
	if err != nil {
		return fmt.Errorf("getting plugin info: %w", err)
	}

	if plugin.DownloadURL == "" {
		return fmt.Errorf("plugin %s has no download URL", domain)
	}

	// Create destination directory
	pluginDir := filepath.Join(destDir, domain)
	if err := os.MkdirAll(pluginDir, 0755); err != nil {
		return fmt.Errorf("creating plugin directory: %w", err)
	}

	// Download the archive
	resp, err := d.client.HTTPClient.Get(plugin.DownloadURL)
	if err != nil {
		return fmt.Errorf("downloading plugin: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("download failed with status %d", resp.StatusCode)
	}

	// Read the full body for checksum verification
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("reading download: %w", err)
	}

	// Verify checksum if provided
	if plugin.Checksum != "" {
		hash := sha256.Sum256(body)
		got := hex.EncodeToString(hash[:])
		if got != plugin.Checksum {
			os.RemoveAll(pluginDir)
			return fmt.Errorf("checksum mismatch: expected %s, got %s", plugin.Checksum, got)
		}
	}

	// Determine archive type and extract
	bodyReader := bytes.NewReader(body)
	if strings.HasSuffix(plugin.DownloadURL, ".tar.gz") || strings.HasSuffix(plugin.DownloadURL, ".tgz") {
		if err := extractTarGz(bodyReader, pluginDir); err != nil {
			return fmt.Errorf("extracting archive: %w", err)
		}
	} else if strings.HasSuffix(plugin.DownloadURL, ".zip") {
		return fmt.Errorf("zip archives not yet supported")
	} else {
		// Assume it's a single manifest.yaml file
		manifestPath := filepath.Join(pluginDir, "manifest.yaml")
		if err := os.WriteFile(manifestPath, body, 0644); err != nil {
			return fmt.Errorf("saving manifest: %w", err)
		}
	}

	// Verify manifest exists
	manifestPath := filepath.Join(pluginDir, "manifest.yaml")
	if _, err := os.Stat(manifestPath); os.IsNotExist(err) {
		// Clean up and return error
		os.RemoveAll(pluginDir)
		return fmt.Errorf("downloaded plugin does not contain manifest.yaml")
	}

	return nil
}

// extractTarGz extracts a tar.gz archive to the destination directory.
func extractTarGz(reader io.Reader, destDir string) error {
	gzr, err := gzip.NewReader(reader)
	if err != nil {
		return err
	}
	defer gzr.Close()

	tr := tar.NewReader(gzr)

	for {
		header, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		// Sanitize the path to prevent path traversal
		target := filepath.Join(destDir, header.Name)
		cleanDest := filepath.Clean(destDir)
		if target != cleanDest && !strings.HasPrefix(target, cleanDest+string(os.PathSeparator)) {
			return fmt.Errorf("invalid file path: %s", header.Name)
		}

		switch header.Typeflag {
		case tar.TypeDir:
			if err := os.MkdirAll(target, 0755); err != nil {
				return err
			}
		case tar.TypeReg:
			// Ensure parent directory exists
			if err := os.MkdirAll(filepath.Dir(target), 0755); err != nil {
				return err
			}

			outFile, err := os.Create(target)
			if err != nil {
				return err
			}

			// Limit copy to prevent decompression bombs
			limitedReader := io.LimitReader(tr, 10*1024*1024) // 10MB max per file
			if _, err := io.Copy(outFile, limitedReader); err != nil {
				outFile.Close()
				return err
			}
			outFile.Close()
		}
	}

	return nil
}

// DownloadInfo contains information about a download.
type DownloadInfo struct {
	Domain       string
	Version      string
	Size         int64
	DownloadURL  string
	InstalledDir string
}

// GetDownloadInfo returns information about a plugin download without actually downloading.
func (d *Downloader) GetDownloadInfo(domain string, destDir string) (*DownloadInfo, error) {
	plugin, err := d.client.GetPlugin(domain)
	if err != nil {
		return nil, err
	}

	return &DownloadInfo{
		Domain:       plugin.Domain,
		Version:      plugin.Version,
		Size:         plugin.Size,
		DownloadURL:  plugin.DownloadURL,
		InstalledDir: filepath.Join(destDir, domain),
	}, nil
}
