package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

type GitLabAssetsLink struct {
	Name           string `json:"name"`
	DirectAssetURL string `json:"direct_asset_url"`
}

type GitLabRelease struct {
	Name            string `json:"name"`
	TagName         string `json:"tag_name"`
	UpcomingRelease bool   `json:"upcoming_release"`
	Assets          struct {
		Links []GitLabAssetsLink `json:"links"`
	} `json:"assets"`
}

type GitLab struct {
	url       string
	apiURL    string
	token     string
	repo      string
	projectID string
}

func NewGitLab(gitlabURL, token, repo string) *GitLab {
	var projectID string
	// Encode project_id if it is not an integer and not encoded
	if !isNumeric(repo) && !isEncoded(repo) {
		projectID = urlEncode(repo)
	} else {
		projectID = repo
	}

	return &GitLab{
		url:       gitlabURL,
		apiURL:    gitlabURL + "/api/v4",
		token:     token,
		repo:      repo,
		projectID: projectID,
	}
}

func (g *GitLab) GetLatestRelease() (GitLabRelease, error) {
	var release GitLabRelease
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/projects/%s/releases?per_page=1", g.apiURL, g.projectID), nil)
	if err != nil {
		return release, err
	}
	req.Header.Set("PRIVATE-TOKEN", g.token)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return release, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return release, fmt.Errorf("Failed to fetch releases, status code: %d", resp.StatusCode)
	}

	var releases []GitLabRelease
	if err := json.NewDecoder(resp.Body).Decode(&releases); err != nil {
		return release, err
	}

	if len(releases) == 0 {
		return release, fmt.Errorf("No release found")
	}

	return releases[0], nil
}

func (g *GitLab) DownloadReleaseAsset(release GitLabRelease, destDir string) (string, error) {
	var (
		found    bool
		filename string
		url      string
	)
	for _, link := range release.Assets.Links {
		if strings.Contains(link.Name, fmt.Sprintf("%s_%s", runtime.GOOS, runtime.GOARCH)) {
			found = true
			filename = link.Name
			url = link.DirectAssetURL
			break
		}
	}

	if !found {
		return "", fmt.Errorf("No valid asset found")
	}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("PRIVATE-TOKEN", g.token)

	log.Printf("Downloading %s from %s", filename, url)
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("Failed to download %s, status code: %d", url, resp.StatusCode)
	}

	destPath := filepath.Join(destDir, filename)
	file, err := os.Create(destPath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	_, err = io.Copy(file, resp.Body)
	if err != nil {
		return "", err
	}
	log.Printf("Downloaded %s", filename)

	return destPath, nil
}
