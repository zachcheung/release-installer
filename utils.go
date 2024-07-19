package main

import (
	"archive/tar"
	"archive/zip"
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
	".zip",
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
	install := func(name string, r io.Reader, mode os.FileMode) error {
		oldpath := filepath.Join(filepath.Dir(archivePath), filepath.Base(name))
		newpath := filepath.Join(destDir, filepath.Base(name))
		outFile, err := os.Create(oldpath)
		if err != nil {
			return err
		}
		if _, err := io.Copy(outFile, r); err != nil {
			return err
		}
		defer outFile.Close()

		if err := outFile.Chmod(mode); err != nil {
			return err
		}

		if err := os.Rename(oldpath, newpath); err != nil {
			return err
		}

		log.Printf("Installed %s to %s", name, destDir)

		return nil
	}

	lowerSrc := strings.ToLower(archivePath)
	if strings.HasSuffix(lowerSrc, ".tar.gz") || strings.HasSuffix(lowerSrc, ".tgz") {
		// gzip tar
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
				if err := install(header.Name, tr, os.FileMode(header.Mode)); err != nil {
					return err
				}
			}
		}
	} else if strings.HasSuffix(lowerSrc, ".zip") {
		// zip
		r, err := zip.OpenReader(archivePath)
		if err != nil {
			return err
		}
		defer r.Close()

		for _, file := range r.File {
			if fh := file.FileHeader; !fh.FileInfo().IsDir() && fh.Mode()&0111 != 0 {
				rc, err := file.Open()
				if err != nil {
					return err
				}
				defer rc.Close()

				if err := install(fh.Name, rc, os.FileMode(fh.Mode())); err != nil {
					return err
				}
			}
		}
	} else {
		return fmt.Errorf("Unsupported archive file: %s", filepath.Base(archivePath))
	}

	return nil
}

func downloadReleaseAsset(release Release, destDir string) (string, error) {
	var (
		maxWeightAsset Asset
		assets         Assets
		err            error
	)
	for _, asset := range release.Assets {
		if !isIgnoredFile(asset.Name) {
			assets = append(assets, asset)
		}
	}

	if release.AssetPattern == nil {
		maxWeightAsset, err = assets.FindMaxWeightAsset()
		if err != nil {
			return "", err
		}
	} else {
		// match by pattern
		var matchedAssets Assets
		for _, asset := range assets {
			if release.AssetPattern.MatchString(asset.Name) {
				matchedAssets = append(matchedAssets, asset)
			}
		}

		switch len(matchedAssets) {
		case 0:
			return "", fmt.Errorf("No matched asset by pattern in assets")
		case 1:
			maxWeightAsset = matchedAssets[0]
		default:
			return "", fmt.Errorf("Multiple matched assets found by pattern in assets: %s", matchedAssets.JoinName())
		}
	}

	destPath := filepath.Join(destDir, maxWeightAsset.Name)
	if err := download(maxWeightAsset.URL, destPath, release.AuthHeaders); err != nil {
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
	if headers != nil {
		for key, value := range headers {
			req.Header.Set(key, value)
		}
	}

	client := &http.Client{}
	return client.Do(req)
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

func compareVersions(v1, v2 string) int {
	v1 = strings.TrimPrefix(strings.ToLower(v1), "v")
	v2 = strings.TrimPrefix(strings.ToLower(v2), "v")
	parts1 := strings.Split(v1, ".")
	parts2 := strings.Split(v2, ".")

	for i := 0; i < len(parts1) || i < len(parts2); i++ {
		var num1, num2 int

		if i < len(parts1) {
			num1, _ = strconv.Atoi(parts1[i])
		}

		if i < len(parts2) {
			num2, _ = strconv.Atoi(parts2[i])
		}

		if num1 < num2 {
			return -1 // less
		} else if num1 > num2 {
			return 1 // greater
		}
	}

	return 0 // equal
}

func boolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}
