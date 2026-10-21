package acquire

import (
	"archive/zip"
	"bytes"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"text/tabwriter"

	"github.com/mgumz/vopher/pkg/common"
	"github.com/mgumz/vopher/pkg/plugin"
)

// Hoster describes the URL layout of a git hoster (GitHub, Forgejo, GitLab,
// ...). Everything else needed to check a plugin for updates - parsing the
// atom feeds, reading the commit from the archive - is the same for all of
// them.
type Hoster interface {
	// GetRepository splits remote into the repository path and the
	// requested head (branch, tag or commit)
	GetRepository(remote *url.URL) (name, head string)

	// CommitsFeedURL returns the atom feed of the commits on branch
	CommitsFeedURL(repo *url.URL, branch string) url.URL

	// TagsFeedURL returns the atom feed of the tags
	TagsFeedURL(repo *url.URL) url.URL

	// PluginArchiveURL returns the URL vopher fetches plugin p from
	PluginArchiveURL(p *plugin.Plugin) *url.URL

	// TagName returns the tag name of an entry-id of the tags feed. The tag
	// name may contain "/".
	TagName(id string) string
}

// GuessCommitByZIP guesses the commit by the comment in the zip file at
// path. `git archive` puts the commit-id there when it is given a commit or
// tag; the zip files of GitHub, Forgejo and GitLab carry it.
func GuessCommitByZIP(path string) string {
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

// GuessCommitByFile "guesses" the git-commit by the content of the file
// written while extracting the zip (see common.CommitFileName)
func GuessCommitByFile(name, base string) string {
	path := filepath.Join(base, name, common.CommitFileName)
	commit, err := os.ReadFile(path) // #nosec G304
	if err != nil {
		return ""
	}
	if len(commit) != common.Sha1ChecksumLen {
		return ""
	}
	return string(commit)
}

// CheckPlugin checks the given plugin for updates and returns printable
// text as result
func CheckPlugin(hoster Hoster, plugin *plugin.Plugin, base, cacheDir string) string {

	name, head := hoster.GetRepository(plugin.URL)
	altHead := GuessCommitByFile(plugin.Name, base)
	if altHead == "" && cacheDir != "" {
		altHead = GuessCommitByZIP(ArchivePath(cacheDir, hoster.PluginArchiveURL(plugin), ".zip"))
	}
	if altHead == "" {
		altHead = GuessCommitByZIP(filepath.Join(base, plugin.Name+".zip"))
	}

	u := *plugin.URL
	u.Path = name

	feeds := buildFeeds(hoster, &u, plugin.Opts.Branch, head)
	feeds.fetch()

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
