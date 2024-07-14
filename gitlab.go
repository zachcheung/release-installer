package main

import (
	"encoding/json"
	"fmt"
	"net/http"
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
	url         string
	apiURL      string
	token       string
	repo        string
	projectID   string
	authHeaders map[string]string
}

func NewGitLab(gitlabURL, token, repo string) *GitLab {
	var projectID string
	// Encode project_id if it is not an integer and not encoded
	if !isNumeric(repo) && !isEncoded(repo) {
		projectID = urlEncode(repo)
	} else {
		projectID = repo
	}

	authHeaders := make(map[string]string)
	if token != "" {
		authHeaders["PRIVATE-TOKEN"] = token
	}

	return &GitLab{
		url:         gitlabURL,
		apiURL:      gitlabURL + "/api/v4",
		token:       token,
		repo:        repo,
		projectID:   projectID,
		authHeaders: authHeaders,
	}
}

func (g *GitLab) GetLatestRelease() (GitLabRelease, error) {
	var release GitLabRelease
	resp, err := httpGet(fmt.Sprintf("%s/projects/%s/releases?per_page=1", g.apiURL, g.projectID), g.authHeaders)
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
		if strings.Contains(link.Name, fmt.Sprintf("%s_%s", runtime.GOOS, runtime.GOARCH)) && strings.HasSuffix(link.Name, ".tar.gz") {
			found = true
			filename = link.Name
			url = link.DirectAssetURL
			break
		}
	}

	if !found {
		return "", fmt.Errorf("No valid asset found")
	}

	destPath := filepath.Join(destDir, filename)
	if err := download(url, destPath, g.authHeaders); err != nil {
		return "", err
	}

	return destPath, nil
}
