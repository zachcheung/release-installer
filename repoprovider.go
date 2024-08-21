package main

import (
	"cmp"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"regexp"
	"runtime"
	"slices"
	"strings"
)

var (
	ErrNoRelease              = errors.New("no release found")
	ErrNoAsset                = errors.New("no asset found")
	ErrMultipleMaxWeightAsset = errors.New("multiple max weight assets")
)

type Asset struct {
	Name                   string
	URL                    string
	containsOS             bool
	containsArch           bool
	matchesOS              bool
	matchesArch            bool
	supportedArchiveFormat bool
	libc                   int
}

func NewAsset(name, url string) *Asset {
	return newAsset(name, url, runtime.GOOS, runtime.GOARCH, isMusl)
}

func newAsset(name, url, goos, goarch string, isMusl bool) *Asset {
	var libc int

	if !isMusl {
		if !containsMusl(name) {
			libc = 1
		} else {
			libc = 0
		}
	} else {
		if containsMusl(name) {
			libc = 1
		} else {
			libc = 0
		}
	}

	return &Asset{
		Name:                   name,
		URL:                    url,
		containsOS:             containsOS(name),
		containsArch:           containsArch(name),
		matchesOS:              matchesOS(name, goos),
		matchesArch:            matchesArch(name, goarch),
		supportedArchiveFormat: isSupportedArchiveFormat(name),
		libc:                   libc,
	}
}

func (a Asset) Weight() int {
	var sum int
	for _, b := range []bool{
		a.containsOS,
		a.containsArch,
		a.matchesOS,
		a.matchesArch,
		a.supportedArchiveFormat,
	} {
		sum += boolToInt(b)
	}
	sum += a.libc
	return sum
}

type Assets []Asset

func (as Assets) FindMaxWeightAsset() (Asset, error) {
	if len(as) == 0 {
		return Asset{}, ErrNoAsset
	}
	// desc
	slices.SortStableFunc(as, func(a, b Asset) int {
		return cmp.Compare(b.Weight(), a.Weight())
	})

	maxWeight := as[0].Weight()
	for i := 1; i < len(as); i++ {
		if as[i].Weight() == maxWeight {
			var maxAs Assets
			for _, v := range as {
				if v.Weight() == maxWeight {
					maxAs = append(maxAs, v)
				} else {
					break
				}
			}
			return Asset{}, fmt.Errorf("%w: %s", ErrMultipleMaxWeightAsset, maxAs.JoinNameWithWeight())
		} else {
			break
		}
	}

	return as[0], nil
}

func (as Assets) JoinName() string {
	return as.joinName(false)
}

func (as Assets) JoinNameWithWeight() string {
	return as.joinName(true)
}

func (as Assets) joinName(withWeight bool) string {
	names := make([]string, len(as))
	for i, asset := range as {
		if withWeight {
			names[i] = fmt.Sprintf("%s: %d", asset.Name, asset.Weight())
		} else {
			names[i] = asset.Name
		}
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
