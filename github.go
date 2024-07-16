package main

import (
	"fmt"
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
	// https://docs.github.com/en/rest/releases/releases?apiVersion=2022-11-28#get-the-latest-release
	url := fmt.Sprintf("https://api.github.com/repos/%s/releases/latest", g.repo)
	return g.getRelease(url)
}

func (g *GitHub) getRelease(url string) (Release, error) {
	var gr GitHubRelease
	if err := GetRelease(url, g.authHeaders, &gr); err != nil {
		return Release{}, err
	}

	return g.convertRelease(gr), nil
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
