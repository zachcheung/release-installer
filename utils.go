package main

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

var supportedArchiveFormat = []string{
	".tar.gz",
	".tgz",
}

func isNumeric(s string) bool {
	_, err := strconv.Atoi(s)
	return err == nil
}

func isEncoded(s string) bool {
	re := regexp.MustCompile(`%[0-9A-Fa-f]{2}`)
	return re.MatchString(s)
}

func urlEncode(s string) string {
	return url.PathEscape(s)
}

func extractAndInstallExecutables(archivePath, destDir string) error {
	file, err := os.Open(archivePath)
	if err != nil {
		return err
	}
	defer file.Close()

	gzr, err := gzip.NewReader(file)
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

		// Check if it is regulare executable file
		if header.Typeflag == tar.TypeReg && header.Mode&0111 != 0 {
			name := header.Name
			outFile, err := os.Create(filepath.Join(destDir, filepath.Base(name)))
			if err != nil {
				return err
			}
			if _, err := io.Copy(outFile, tr); err != nil {
				return err
			}
			defer outFile.Close()

			if err := outFile.Chmod(os.FileMode(header.Mode)); err != nil {
				return err
			}

			log.Printf("Installed %s to %s", name, destDir)
		}
	}

	return nil
}

func downloadReleaseAsset(release Release, destDir string) (string, error) {
	var (
		matchedAsset  Asset
		matchedAssets []Asset
	)
	for _, asset := range release.Assets {
		if matchAsset(asset.Name) {
			matchedAssets = append(matchedAssets, asset)
		}
	}

	switch len(matchedAssets) {
	case 0:
		if len(release.Assets) == 1 {
			// iamseth/oracledb_exporter: oracledb_exporter.tar.gz
			matchedAsset = release.Assets[0]
		} else {
			return "", fmt.Errorf("No valid asset found")
		}
	case 1:
		matchedAsset = matchedAssets[0]
	default:
		var found bool
		for _, asset := range matchedAssets {
			if isSupportedArchiveFormat(asset.Name) {
				found = true
				matchedAsset = asset
				break
			}
		}

		if !found {
			return "", fmt.Errorf("No valid asset found in matched assets")
		}
	}

	destPath := filepath.Join(destDir, matchedAsset.Name)
	if err := download(matchedAsset.URL, destPath, release.AuthHeaders); err != nil {
		return "", err
	}

	return destPath, nil
}

func download(url, destPath string, headers map[string]string) error {
	filename := filepath.Base(destPath)
	log.Printf("Downloading %s from %s", filename, url)
	resp, err := httpGet(url, headers)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("Failed to download %s, status code: %d", url, resp.StatusCode)
	}

	if err := copyReader(destPath, resp.Body); err != nil {
		return err
	}
	log.Printf("Downloaded %s", filename)

	return nil
}

func httpGet(url string, headers map[string]string) (*http.Response, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	for key, value := range headers {
		req.Header.Set(key, value)
	}

	client := &http.Client{}
	return client.Do(req)
}

func copyFile(dst, src string) error {
	srcFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	return copyReader(dst, srcFile)
}

func copyReader(dst string, src io.Reader) error {
	file, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = io.Copy(file, src)
	if err != nil {
		return err
	}

	return nil
}

func addExecutePermission(fpath string) error {
	file, err := os.Open(fpath)
	if err != nil {
		return fmt.Errorf("failed to open file: %v", err)
	}
	defer file.Close()

	info, err := file.Stat()
	if err != nil {
		return fmt.Errorf("Failed to get file info: %v", err)
	}

	currentMode := info.Mode()
	newMode := currentMode | 0111
	if err := file.Chmod(newMode); err != nil {
		return fmt.Errorf("Failed to change file mode: %v", err)
	}

	return nil
}

func isSupportedArchiveFormat(name string) bool {
	name = strings.ToLower(name)
	for _, suffix := range supportedArchiveFormat {
		if strings.HasSuffix(name, suffix) {
			return true
		}
	}
	return false
}
