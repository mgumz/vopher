package acquire

import (
	"archive/zip"
	"bytes"
	"fmt"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strings"
	"text/tabwriter"

	"github.com/mgumz/vopher/pkg/common"
	"github.com/mgumz/vopher/pkg/plugin"
	"github.com/mgumz/vopher/pkg/utils"
)

const (
	nGithubRepoParts = 3
)

// Github is a pseudo type to group Github related functions
type Github struct{}

// GetRepository extracts parts of `url`. The repo-name is usually the first 2
// parts of the repo.Path:
//
//	github.com/username/reponame
//	github.com/username/reponame/archive/master.zip
//	github.com/username/reponame/archive/main.zip
//	github.com/username/reponame#v2.1
func (gh Github) GetRepository(remote *url.URL) (name, head string) {
	name = remote.Path
	if idx := utils.IndexByteN(remote.Path, '/', nGithubRepoParts); idx > 0 {
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

	if len(zfile.Comment) == common.Sha1ChecksumLen {
		return zfile.Comment
	}
	return ""
}

// GuessCommitByFile "guesses" the git-commit by the content of file named
// "github-commit"
func (gh Github) GuessCommitByFile(name, base string) string {
	path := filepath.Join(base, name, "github-commit")
	commit, err := os.ReadFile(path) // #nosec G304
	if err != nil {
		return ""
	}
	if len(commit) != common.Sha1ChecksumLen {
		return ""
	}
	return string(commit)
}

// FeedURL builds an URL from the base url of repo and the given parts
func (gh Github) FeedURL(repo *url.URL, parts ...string) url.URL {
	feedURL := *repo
	feedURL.Path = path.Join(feedURL.Path, path.Join(parts...)) + ".atom"
	return feedURL
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

	feeds := buildFeeds(plugin.Opts.Branch, head)
	feeds.fetch(gh, &u)

	buf := bytes.NewBuffer(nil)

	// output header
	const foldMarkerBegin = "  ✚"
	const foldMarkerEnd = "▲"
	fmt.Fprintf(buf, "## %s - %s%s\n", plugin.Name, plugin.URL, foldMarkerBegin)

	// output body
	tw := tabwriter.NewWriter(buf, 0, 0, 1, ' ', 0)
	for i := range feeds {
		feed := feeds[i].feed
		section := feeds[i].section

		if feed == nil || len(feed.Entry) == 0 {
			continue
		}

		fmt.Fprintln(tw, section)
		for _, entry := range feed.Entry {
			mark := " "
			if entry.ID == head || entry.ID == altHead {
				mark = "*"
			}
			title := strings.ReplaceAll(entry.Title, "\t", " ")
			fmt.Fprintf(tw, "  %s%s\t%s\t%s\n", mark, shortID(entry.ID), entry.Updated, title)
		}
	}
	tw.Flush()
	fmt.Fprintf(buf, "\n%s\n", foldMarkerEnd)

	return buf.String()
}

// shortID cuts a commit-id to 10 chars, like `git log --oneline`. Tag names
// are kept as they are.
func shortID(id string) string {
	const n = 10
	if len(id) == common.Sha1ChecksumLen && strings.Trim(id, "0123456789abcdef") == "" {
		return id[:n]
	}
	return id
}
