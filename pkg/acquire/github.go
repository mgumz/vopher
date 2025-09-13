package acquire

import (
	"archive/zip"
	"bytes"
	"encoding/xml"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strings"
	"sync"

	"github.com/mgumz/vopher/pkg/plugin"
	"github.com/mgumz/vopher/pkg/utils"
)

const (
	sha1CheckSumLen = 40
)

// Github is a pseudo type to group Github related functions
type Github struct{}

// GetRepository extracts parts of `url`. The repo-name is usually the first 2
// parts of the repo.Path:
//
//	github.com/username/reponame
//	github.com/username/reponame/archive/master.zip
//	github.com/username/reponame#v2.1
func (gh Github) GetRepository(remote *url.URL) (name, head string) {
	name = remote.Path
	if idx := utils.IndexByteN(remote.Path, '/', 3); idx > 0 {
		name = name[:idx]
	}

	// TODO: other means to detect the current used 'head'
	head = "master"
	if remote.Fragment != "" {
		head = remote.Fragment
	} else if strings.HasSuffix(remote.Path, ".zip") {
		head = path.Base(remote.Path)
		head = head[:len(head)-4]
	}

	return name, head
}

// GuessCommitByZIP guesses the commit by the comment in a github-zip file. It
// refers to the git-commit-id used by github to create the zip.
func (gh Github) GuessCommitByZIP(name, base string) string {
	path := filepath.Join(base, name+".zip")
	zfile, err := zip.OpenReader(path)
	if err != nil {
		return ""
	}
	defer zfile.Close()

	if len(zfile.Comment) == sha1CheckSumLen {
		return zfile.Comment
	}
	return ""
}

// GuessCommitByFile "guesses" the git-commit by the content of file named
// "github-commit"
func (gh Github) GuessCommitByFile(name, base string) string {
	path := filepath.Join(base, name, "github-commit")
	commit, err := os.ReadFile(path)
	if err != nil {
		return ""
	}
	if len(commit) != sha1CheckSumLen {
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

// FeedURL builds an URL from the base url of repo and the given parts
func (gh Github) FeedURL(repo *url.URL, parts ...string) url.URL {
	feedURL := *repo
	feedURL.Path = path.Join(feedURL.Path, path.Join(parts...)) + ".atom"
	return feedURL
}

// GetCommits returns the GithubFeed of the repository
func (gh Github) GetCommits(repo *url.URL, parts ...string) *GithubFeed {

	feedURL := gh.FeedURL(repo, parts...)
	resp, err := http.Get(feedURL.String())
	if err != nil {
		log.Printf("error: %q %v", feedURL.String(), err)
		return nil
	}
	defer (func() { _ = resp.Body.Close() })()

	if resp.StatusCode == http.StatusNotFound {
		return nil
	}
	if resp.StatusCode != http.StatusOK {
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

// CheckPlugin checks the given plugin for updates and returns printable
// text as result
func (gh Github) CheckPlugin(plugin *plugin.Plugin, base string) string {

	name, head := gh.GetRepository(plugin.URL)
	altHead := gh.GuessCommitByFile(plugin.Name, base)
	if altHead == "" {
		altHead = gh.GuessCommitByZIP(plugin.Name, base)
	}

	u := *plugin.URL
	u.Path = name

	commits := []struct {
		parts   []string
		feed    *GithubFeed
		section string
	}{
		{[]string{"commits", "master"}, nil, "\n - master commits:\n"},
		{[]string{"tags"}, nil, "\n - tags:\n"},
	}
	if head != "master" {
		commit := commits[0]
		commit.parts = []string{"commits", head}
		commit.section = "\n - commits:\n"
		commits = append(commits, commit)
	}

	var wg sync.WaitGroup
	wg.Add(len(commits))
	for i := range commits {
		go func(j int) {
			commits[j].feed = gh.GetCommits(&u, commits[j].parts...)
			wg.Done()
		}(i)
	}
	wg.Wait()

	buf := bytes.NewBuffer(nil)

	fmt.Fprintf(buf, "\n\n## %s - %s\n", plugin.Name, plugin.URL)

	for i := range commits {
		c := &commits[i]
		if c.feed == nil || len(c.feed.Entry) == 0 {
			continue
		}

		fmt.Fprintln(buf, c.section)
		for _, entry := range c.feed.Entry {
			mark := " "
			if entry.ID == head || entry.ID == altHead {
				mark = "*"
			}
			fmt.Fprintf(buf, "  %s%.10s %s %s\n", mark, entry.ID, entry.Updated, entry.Title)
		}
	}

	return buf.String()
}
