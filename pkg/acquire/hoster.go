package acquire

import (
	"archive/zip"
	"bytes"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strings"
	"text/tabwriter"

	"github.com/mgumz/vopher/pkg/archive"
	"github.com/mgumz/vopher/pkg/common"
	vhttp "github.com/mgumz/vopher/pkg/http"
	"github.com/mgumz/vopher/pkg/plugin"
	"github.com/mgumz/vopher/pkg/utils"
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

	// ArchiveURL returns the zip archive of ref (branch, tag or commit)
	ArchiveURL(repo *url.URL, ref string) url.URL

	// TagsFeedURL returns the atom feed of the tags
	TagsFeedURL(repo *url.URL) url.URL

	// BranchOf returns the branch holding commit. If several do, it prefers
	// branch (`branch=`, may be ""), then the default branches, see
	// branchCandidates. It returns "" if it finds none.
	BranchOf(repo *url.URL, commit, branch string) (string, error)

	// TagName returns the tag name of an entry-id of the tags feed. The tag
	// name may contain "/".
	TagName(id string) string
}

// PluginArchiveURLs returns the URLs vopher may have fetched plugin p from:
// p.URL if it already points to an archive, else the archive of the ref
// given by the fragment of p.URL or `branch=` (in that order). Without
// both, `update` probes DefaultBranches, so it is the archive of each of
// them.
func PluginArchiveURLs(hoster Hoster, p *plugin.Plugin) []*url.URL {
	if ok, _ := archive.IsSupportedArchive(path.Base(p.URL.Path)); ok {
		return []*url.URL{p.URL}
	}
	refs := DefaultBranches
	if ref := utils.FirstNotEmpty(p.URL.Fragment, p.Opts.Branch); ref != "" {
		refs = []string{ref}
	}
	urls := []*url.URL{}
	for _, ref := range refs {
		u := hoster.ArchiveURL(p.URL, ref)
		urls = append(urls, &u)
	}
	return urls
}

// hosters maps the names usable in `hoster=` to the Hoster
var hosters = map[string]Hoster{
	"github":  Github{},
	"forgejo": Forgejo{},
	"gitea":   Forgejo{},
	"gitlab":  Gitlab{},
}

// HosterFor returns the Hoster of plugin: the one named by `hoster=`, else
// the one of the host of the plugin URL. It returns nil for hosts it does
// not know.
func HosterFor(plugin *plugin.Plugin) (Hoster, error) {
	name := plugin.Opts.Hoster
	if name == "" {
		name = common.KnownHosts[strings.ToLower(plugin.URL.Hostname())]
	}
	if name == "" {
		return nil, nil
	}
	hoster, ok := hosters[name]
	if !ok {
		return nil, fmt.Errorf("unknown hoster %q", name)
	}
	return hoster, nil
}

// DefaultBranches are probed by ProbeBranch, in this order, for plugins
// without `#ref` and without `branch=`
var DefaultBranches = []string{"main", "master"}

// ProbeBranch returns the first of branches with an archive at the hoster.
// It asks with HEAD requests.
func ProbeBranch(hoster Hoster, repo *url.URL, branches ...string) (string, error) {
	for _, branch := range branches {
		archiveURL := hoster.ArchiveURL(repo, branch)
		resp, done, err := vhttp.Request(http.MethodHead, archiveURL.String())
		if err != nil {
			return "", err
		}
		done()
		if resp.StatusCode == http.StatusOK {
			return branch, nil
		}
	}
	return "", fmt.Errorf("no archive for any of the branches %q", branches)
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
// written while extracting the zip (see common.CommitFileName), or of the
// file written by older versions of vopher
func GuessCommitByFile(name, base string) string {
	path := filepath.Join(base, name, common.CommitFileName)
	commit, err := os.ReadFile(path) // #nosec G304
	if os.IsNotExist(err) {
		path = filepath.Join(base, name, common.LegacyCommitFileName)
		commit, err = os.ReadFile(path) // #nosec G304
	}
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
		for _, u := range PluginArchiveURLs(hoster, plugin) {
			if altHead = GuessCommitByZIP(ArchivePath(cacheDir, u, ".zip")); altHead != "" {
				break
			}
		}
	}
	if altHead == "" {
		altHead = GuessCommitByZIP(filepath.Join(base, plugin.Name+".zip"))
	}

	u := *plugin.URL
	u.Path = name

	// a #ref: show the commits of the branch holding the installed commit
	refBranch := ""
	if plugin.URL.Fragment != "" && altHead != "" {
		var err error
		if refBranch, err = hoster.BranchOf(&u, altHead, plugin.Opts.Branch); err != nil {
			log.Printf("error: %q %v", plugin.URL, err)
		}
	}

	feeds := buildFeeds(hoster, &u, plugin.Opts.Branch, head, refBranch)
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

// IsCommit returns true if ref looks like a (short) git commit id. The
// archive of a commit never changes, unlike the one of a branch or tag.
func IsCommit(ref string) bool {
	const minShortCommitLen = 7
	if len(ref) < minShortCommitLen || len(ref) > common.Sha1ChecksumLen {
		return false
	}
	for _, c := range ref {
		if !strings.ContainsRune("0123456789abcdefABCDEF", c) {
			return false
		}
	}
	return true
}

// ownerRepository extracts parts of `url` for hosters with repositories at
// "<owner>/<repo>" (GitHub, Forgejo). The repo-name is the first 2 parts of
// the repo.Path; head is the #fragment, the name of the .zip or "master".
func ownerRepository(remote *url.URL) (name, head string) {
	const nRepoParts = 3
	name = remote.Path
	if idx := utils.IndexByteN(remote.Path, '/', nRepoParts); idx > 0 {
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
