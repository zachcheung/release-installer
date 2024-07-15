package main

import (
	"encoding/json"
	"fmt"
	"net/http"
)

type GitHubAsset struct {
	Name               string `json:"name"`
	URL                string `json:"url"`
	BrowserDownloadURL string `json:"browser_download_url"`
}

type GitHubRelease struct {
	Name       string        `json:"name"`
	TagName    string        `json:"tag_name"`
	Draft      bool          `json:"draft"`
	Prerelease bool          `json:"prerelease"`
	Assets     []GitHubAsset `json:"assets"`
}

type GitHub struct {
	token       string
	repo        string
	authHeaders map[string]string
}

func NewGitHub(token, repo string) *GitHub {
	authHeaders := make(map[string]string)
	if token != "" {
		authHeaders["Authorization"] = "Bearer " + token
	}
	return &GitHub{
		token:       token,
		repo:        repo,
		authHeaders: authHeaders,
	}
}

func (g *GitHub) GetLatestRelease() (Release, error) {
	var release Release
	url := fmt.Sprintf("https://api.github.com/repos/%s/releases?per_page=1", g.repo)
	resp, err := httpGet(url, g.authHeaders)
	if err != nil {
		return release, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return release, fmt.Errorf("Failed to fetch releases, status code: %d, url: %s", resp.StatusCode, url)
	}

	var releases []GitHubRelease
	if err := json.NewDecoder(resp.Body).Decode(&releases); err != nil {
		return release, err
	}

	if len(releases) == 0 {
		return release, fmt.Errorf("No release found")
	}

	return g.convertRelease(releases[0]), nil
}

func (g *GitHub) convertRelease(gr GitHubRelease) Release {
	// https://docs.github.com/en/rest/releases/assets?apiVersion=2022-11-28#get-a-release-asset
	headers := g.authHeaders
	if g.token != "" {
		headers["Accept"] = "application/octet-stream"
	}
	r := Release{
		Name:        gr.Name,
		TagName:     gr.TagName,
		AuthHeaders: headers,
	}
	for _, ga := range gr.Assets {
		var url string
		if g.token == "" {
			url = ga.BrowserDownloadURL
		} else {
			url = ga.URL
		}
		asset := Asset{
			Name: ga.Name,
			URL:  url,
		}
		r.Assets = append(r.Assets, asset)
	}
	return r
}