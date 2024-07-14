package main

type Asset struct {
	Name string
	URL  string
}

type Release struct {
	Name        string
	TagName     string
	Assets      []Asset
	AuthHeaders map[string]string
}

type RepoProvider interface {
	GetLatestRelease() (Release, error)
}
