package main

import (
	"archive/tar"
	"compress/gzip"
	"io"
	"log"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
)

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
