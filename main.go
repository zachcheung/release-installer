package main

import (
	"flag"
	"fmt"
	"log"
	"os"
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
	flag.StringVar(&provider, "provider", "gitlab", "repo provider, options: gitlab")
	flag.StringVar(&baseURL, "url", "", "base url")
	flag.StringVar(&token, "token", "", "token for private repo")
	flag.Parse()

	if flag.NArg() == 0 {
		fmt.Println("Missing repo")
		os.Exit(1)
	}
	repo = flag.Arg(0)

	var g RepoProvider
	switch provider {
	case "gitlab":
		g = NewGitLab(baseURL, token, repo)
	default:
		fmt.Printf("unsupported provider: %s\n", provider)
		os.Exit(1)
	}

	release, err := g.GetLatestRelease()
	if err != nil {
		log.Fatalf("Error getting latest release: %v", err)
	}

	tempDir, err := os.MkdirTemp("", "release-installer")
	if err != nil {
		log.Fatalf("Error creating temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	archivePath, err := downloadReleaseAsset(release, tempDir)
	if err != nil {
		log.Fatalf("Error downloading asset: %v", err)
	}
	if err := extractAndInstallExecutables(archivePath, installDir); err != nil {
		log.Fatalf("Error installing package: %v", err)
	}
}
