package main

import (
	"errors"
	"strings"
)

var ErrNoRelease = errors.New("No release found")

type Asset struct {
	Name string
	URL  string
}

type Assets []Asset

func (as Assets) JoinName() string {
	names := make([]string, len(as))
	for i, asset := range as {
		names[i] = asset.Name
	}
	return strings.Join(names, ", ")
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
