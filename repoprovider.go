package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"regexp"
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
	Name         string
	TagName      string
	Assets       []Asset
	AuthHeaders  map[string]string
	AssetPattern *regexp.Regexp
}

type RepoProvider interface {
	GetLatestRelease() (Release, error)
	GetTaggedRelease(tag string) (Release, error)
}

func GetRelease(url string, headers map[string]string, target interface{}) error {
	resp, err := httpGet(url, headers)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		if resp.StatusCode == http.StatusNotFound {
			return ErrNoRelease
		}
		return fmt.Errorf("failed to fetch release, status code: %d, url: %s", resp.StatusCode, url)
	}

	if err := json.NewDecoder(resp.Body).Decode(target); err != nil {
		return err
	}

	return nil
}
