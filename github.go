package main

import (
	"archive/zip"
	"encoding/xml"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"path"
	"path/filepath"
	"strings"
)

// Github is a pseudo type to group Github related functions
type Github struct{}

// the repo-name is usually the first 2 parts of the
// repo.Path:
//    github.com/username/reponame
//    github.com/username/reponame/archive/master.zip
//    github.com/username/reponame#v2.1
func (gh Github) getRepository(remote *url.URL) (name, head string) {
	name = remote.Path
	if idx := indexByteN(remote.Path, '/', 3); idx > 0 {
		name = name[:idx]
	}

	// TODO: other means to detect the current used 'head'
	//
	head = "master"
	if remote.Fragment != "" {
		head = remote.Fragment
	} else if strings.HasSuffix(remote.Path, ".zip") {
		head = path.Base(remote.Path)
		head = head[:len(head)-4]
	}

	return name, head
}

// the comment in a github-zip file refers to the git-commit-id
// used by github to create the zip.
func (gh Github) guessCommitByZIP(name, base string) string {
	path := filepath.Join(base, name+".zip")
	zfile, err := zip.OpenReader(path)
	if err != nil {
		return ""
	}
	defer zfile.Close()

	if len(zfile.Comment) == 40 { // FIXME: magic constant
		return zfile.Comment
	}
	return ""
}

func (gh Github) guessCommitByFile(name, base string) string {
	path := filepath.Join(base, name, "github-commit")
	commit, err := ioutil.ReadFile(path)
	if err != nil {
		return ""
	}
	if len(commit) != 40 {
		return ""
	}
	return string(commit)
}

// GithubFeed is a minimal atom-parser sufficient to extract only what we need
// from Github
type GithubFeed struct {
	XMLName xml.Name `xml:"feed"`
	Entry   []struct {
		Title   string `xml:"title"`
		Updated string `xml:"updated"`
		ID      string `xml:"id"`
	} `xml:"entry"`
}

func (gh Github) feedURL(repo *url.URL, parts ...string) url.URL {
	feedURL := *repo
	feedURL.Path = path.Join(feedURL.Path, path.Join(parts...)) + ".atom"
	return feedURL
}

func (gh Github) getCommits(repo *url.URL, parts ...string) *GithubFeed {

	feedURL := gh.feedURL(repo, parts...)
	resp, err := http.Get(feedURL.String())
	if err != nil {
		log.Printf("error: %q %v", feedURL.String(), err)
		return nil
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		log.Printf("error: %q %v", feedURL.String(), err)
		return nil
	}

	xmldec := xml.NewDecoder(resp.Body)
	feed := new(GithubFeed)

	if err = xmldec.Decode(feed); err != nil {
		log.Printf("error: %q %v", feedURL.String(), err)
		return nil
	}

	if len(feed.Entry) == 0 {
		return nil
	}

	for i, entry := range feed.Entry {
		idx := strings.LastIndex(entry.ID, "/")
		id := entry.ID[idx+1:]
		title := strings.TrimSpace(entry.Title)
		feed.Entry[i].ID = id
		feed.Entry[i].Title = title
	}

	return feed
}
