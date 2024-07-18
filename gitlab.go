package main

import (
	"fmt"
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

func (g *GitLab) GetLatestRelease() (Release, error) {
	// https://docs.gitlab.com/ee/api/releases/#get-the-latest-release
	url := fmt.Sprintf("%s/projects/%s/releases/permalink/latest", g.apiURL, g.projectID)
	return g.getRelease(url)
}

func (g *GitLab) GetTaggedRelease(tag string) (Release, error) {
	// https://docs.gitlab.com/ee/api/releases/#get-a-release-by-a-tag-name
	url := fmt.Sprintf("%s/projects/%s/releases/%s", g.apiURL, g.projectID, tag)
	return g.getRelease(url)
}

func (g *GitLab) getRelease(url string) (Release, error) {
	// https://docs.gitlab.com/ee/api/releases/#get-the-latest-release
	var gr GitLabRelease
	if err := GetRelease(url, g.authHeaders, &gr); err != nil {
		return Release{}, err
	}

	return g.convertRelease(gr), nil
}

func (g *GitLab) convertRelease(gr GitLabRelease) Release {
	r := Release{
		Name:        gr.Name,
		TagName:     gr.TagName,
		AuthHeaders: g.authHeaders,
	}
	for _, link := range gr.Assets.Links {
		r.Assets = append(r.Assets, *NewAsset(link.Name, link.DirectAssetURL))
	}
	return r
}
