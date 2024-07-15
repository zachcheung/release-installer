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
	"runtime"
	"strconv"
	"strings"
)

var packageRe = regexp.MustCompile(fmt.Sprintf("%s[_-]%s", runtime.GOOS, runtime.GOARCH))

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
			outFile, err := os.Create(filepath.Join(destDir, name))
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
		found    bool
		filename string
		url      string
	)
	for _, asset := range release.Assets {
		name := asset.Name
		if packageRe.MatchString(name) && strings.HasSuffix(name, ".tar.gz") {
			found = true
			filename = name
			url = asset.URL
			break
		}
	}

	if !found {
		return "", fmt.Errorf("No valid asset found")
	}

	destPath := filepath.Join(destDir, filename)
	if err := download(url, destPath, release.AuthHeaders); err != nil {
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

	file, err := os.Create(destPath)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = io.Copy(file, resp.Body)
	if err != nil {
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
