package main

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"slices"
	"strings"

	"golang.org/x/net/html"
)

var ErrNotFound = errors.New("not found")

type ApacheAsset struct {
	Name string
	URL  string
}

type ApacheRelease struct {
	Name    string
	TagName string
	Assets  []ApacheAsset
	URL     string
}

type Apache struct {
	url string
}

type Link struct {
	Name string
	URL  string
}

func NewApache(url string) *Apache {
	return &Apache{
		url: url,
	}
}

func (a *Apache) GetLatestRelease() (Release, error) {
	ars, err := a.getReleases()
	if err != nil {
		return Release{}, err
	}
	if len(ars) == 0 {
		return Release{}, ErrNoRelease
	}

	ar := slices.MaxFunc(ars, func(a, b ApacheRelease) int {
		return compareVersions(a.Name, b.Name)
	})

	return a.getRelease(ar)
}

func (a *Apache) GetTaggedRelease(tag string) (Release, error) {
	ars, err := a.getReleases()
	if err != nil {
		return Release{}, err
	}
	if len(ars) == 0 {
		return Release{}, ErrNoRelease
	}

	i := slices.IndexFunc(ars, func(a ApacheRelease) bool {
		return a.Name == tag
	})
	if i == -1 {
		return Release{}, ErrNoRelease
	}

	ar := ars[i]
	return a.getRelease(ar)
}

func (a *Apache) getReleases() ([]ApacheRelease, error) {
	baseURL := a.url
	links, err := getLinks(baseURL)
	if err != nil {
		return nil, err
	}

	var ars []ApacheRelease
	for _, link := range links {
		name := strings.TrimSuffix(link.Name, "/")
		u, err := url.JoinPath(baseURL, link.URL)
		if err != nil {
			return nil, err
		}
		ar := ApacheRelease{
			Name:    name,
			TagName: name,
			URL:     u,
		}
		ars = append(ars, ar)

	}

	return ars, nil
}

func (a *Apache) getRelease(ar ApacheRelease) (Release, error) {
	baseURL := ar.URL
	links, err := getLinks(baseURL)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return Release{}, ErrNoRelease
		}
		return Release{}, err
	}

	var aas []ApacheAsset
	for _, link := range links {
		name := strings.TrimSuffix(link.Name, "/")
		u, err := url.JoinPath(baseURL, link.URL)
		if err != nil {
			return Release{}, err
		}
		as := ApacheAsset{
			Name: name,
			URL:  u,
		}
		aas = append(aas, as)

	}
	ar.Assets = aas

	return a.convertRelease(ar), nil
}

func (a *Apache) convertRelease(ar ApacheRelease) Release {
	r := Release{
		Name:    ar.Name,
		TagName: ar.TagName,
	}
	for _, aa := range ar.Assets {
		r.Assets = append(r.Assets, *NewAsset(aa.Name, aa.URL))
	}
	return r
}

func getLinks(url string) ([]Link, error) {
	resp, err := httpGet(url, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		if resp.StatusCode == http.StatusNotFound {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("failed to fetch url, status code: %d, url: %s", resp.StatusCode, url)
	}
	var links []Link
	n, err := html.Parse(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("Error parsing HTML: %v", err)
	}

	var f func(*html.Node)
	f = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "a" {
			for _, attr := range n.Attr {
				if attr.Key == "href" && !strings.Contains(attr.Val, "?") {
					if n.FirstChild != nil && n.FirstChild.Type == html.TextNode && n.FirstChild.Data != "Parent Directory" {
						links = append(links, Link{Name: n.FirstChild.Data, URL: attr.Val})
					}
				}
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			f(c)
		}
	}

	f(n)
	return links, nil
}
