package main

import (
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

var (
	installDir   string
	provider     string
	baseURL      string
	token        string
	tag          string
	repo         string
	pattern      string
	printVersion bool
	version      string
)

func main() {
	flag.StringVar(&installDir, "dir", "/usr/local/bin", "installation directory")
	flag.StringVar(&provider, "provider", "", "repo provider, default is github, options: github, gitlab, apache")
	flag.StringVar(&baseURL, "url", "", "base url, e.g., https://gitlab.example.com")
	flag.StringVar(&token, "token", "", "token for private repo")
	flag.StringVar(&tag, "tag", "", "tag name, v can be omitted")
	flag.StringVar(&pattern, "pattern", "", "match asset by regexp")
	flag.BoolVar(&printVersion, "version", false, "print version")
	flag.Parse()

	if printVersion {
		fmt.Println(version)
		return
	}

	if flag.NArg() == 0 {
		fmt.Println("Missing repo")
		os.Exit(1)
	}
	repo = flag.Arg(0)

	var (
		patternRe *regexp.Regexp
		err       error
	)
	if pattern != "" {
		if patternRe, err = regexp.Compile(pattern); err != nil {
			log.Fatalf("Invalid pattern: %v", err)
		}
	}

	if err := os.MkdirAll(installDir, 0755); err != nil {
		log.Fatalf("Error creating directory: %v", err)
	}

	if provider == "gitlab" && baseURL == "" {
		baseURL = "https://gitlab.com"
	}
	if provider == "" && strings.Contains(strings.ToLower(baseURL), "gitlab") {
		provider = "gitlab"
	}
	if provider == "apache" && baseURL == "" {
		log.Fatalf("-url is required with apache provider")
	}
	if provider == "" {
		provider = "github"
	}

	var g RepoProvider
	switch provider {
	case "github":
		g = NewGitHub(token, repo)
	case "gitlab":
		g = NewGitLab(baseURL, token, repo)
	case "apache":
		g = NewApache(baseURL)
	default:
		log.Fatalf("unsupported provider: %s", provider)
	}

	var release Release
	if tag == "" {
		release, err = g.GetLatestRelease()
	} else {
		release, err = g.GetTaggedRelease(tag)
		if err != nil && errors.Is(err, ErrNoRelease) && !strings.HasPrefix(tag, "v") {
			// try again with v prefix
			vTag := "v" + tag
			release, err = g.GetTaggedRelease(vTag)
		}
	}
	if err != nil {
		if errors.Is(err, ErrNoRelease) {
			log.Print("No release found")
			return
		} else {
			log.Fatalf("Error getting release: %v", err)
		}
	}

	if len(release.Assets) == 0 {
		log.Print("Empty release assets")
		return
	}

	if patternRe != nil {
		release.AssetPattern = patternRe
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
		isSameFile, err := isIdenticalFile(fpath, destPath)
		if err != nil {
			log.Fatal(err)
		}
		if isSameFile {
			log.Printf("%s is identical, no need to install", destPath)
			return
		}

		if err := addExecutePermission(fpath); err != nil {
			log.Fatalf("Error adding execute permission: %v", err)
		}
		if err := os.Rename(fpath, destPath); err != nil {
			log.Fatalf("Error installing package: %v", err)
		}
		log.Printf("Installed %s as %s", filepath.Base(fpath), destPath)
	}
}
