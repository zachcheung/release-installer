package main

import (
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
)

var (
	installDir string
	provider   string
	baseURL    string
	token      string
	repo       string
)

func main() {
	flag.StringVar(&installDir, "dir", "/usr/local/bin", "installation directory")
	flag.StringVar(&provider, "provider", "github", "repo provider, options: github, gitlab")
	flag.StringVar(&baseURL, "url", "", "base url, e.g., https://gitlab.example.com")
	flag.StringVar(&token, "token", "", "token for private repo")
	flag.Parse()

	if flag.NArg() == 0 {
		fmt.Println("Missing repo")
		os.Exit(1)
	}
	repo = flag.Arg(0)

	if err := os.MkdirAll(installDir, 0755); err != nil {
		log.Fatalf("Error creating directory: %v", err)
	}

	if provider == "gitlab" && baseURL == "" {
		baseURL = "https://gitlab.com"
	}
	if strings.Contains(strings.ToLower(baseURL), "gitlab") && provider == "github" {
		provider = "gitlab"
	}

	var g RepoProvider
	switch provider {
	case "github":
		g = NewGitHub(token, repo)
	case "gitlab":
		g = NewGitLab(baseURL, token, repo)
	default:
		fmt.Printf("unsupported provider: %s\n", provider)
		os.Exit(1)
	}

	release, err := g.GetLatestRelease()
	if err != nil {
		if errors.Is(err, ErrNoRelease) {
			log.Print("No release found")
			return
		} else {
			log.Fatalf("Error getting latest release: %v", err)
		}
	}

	if len(release.Assets) == 0 {
		log.Print("Empty release assets")
		return
	}

	tempDir, err := os.MkdirTemp("", "release-installer")
	if err != nil {
		log.Fatalf("Error creating temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	fpath, err := downloadReleaseAsset(release, tempDir)
	if err != nil {
		log.Fatalf("Error downloading asset: %v", err)
	}

	if isSupportedArchiveFormat(fpath) {
		if err := extractAndInstallExecutables(fpath, installDir); err != nil {
			log.Fatalf("Error installing package: %v", err)
		}
	} else {
		// use repo base as filename
		name := filepath.Base(repo)
		destPath := filepath.Join(installDir, name)
		if err := copyFile(destPath, fpath); err != nil {
			log.Fatalf("Error installing package: %v", err)
		}
		if err := addExecutePermission(destPath); err != nil {
			log.Fatalf("Error adding execute permission: %v", err)
		}
		log.Printf("Installed %s as %s", filepath.Base(fpath), filepath.Join(installDir, name))
	}
}
