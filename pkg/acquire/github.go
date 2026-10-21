package acquire

import (
	"net/url"
	"path"
	"strings"

	"github.com/mgumz/vopher/pkg/archive"
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

// CommitsFeedURL returns github.com/username/reponame/commits/<branch>.atom
func (gh Github) CommitsFeedURL(repo *url.URL, branch string) url.URL {
	return gh.feedURL(repo, "commits", branch)
}

// TagsFeedURL returns github.com/username/reponame/tags.atom
func (gh Github) TagsFeedURL(repo *url.URL) url.URL {
	return gh.feedURL(repo, "tags")
}

func (gh Github) feedURL(repo *url.URL, parts ...string) url.URL {
	feedURL := *repo
	feedURL.Path = path.Join(feedURL.Path, path.Join(parts...)) + ".atom"
	return feedURL
}

// TagName returns <tag> of "tag:github.com,2008:Repository/<repo-id>/<tag>"
func (gh Github) TagName(id string) string {
	if _, rest, ok := strings.Cut(id, ":Repository/"); ok {
		if _, tag, ok := strings.Cut(rest, "/"); ok {
			return tag
		}
	}
	return commitID(id)
}

// ArchiveURL returns the URL of the zip github creates for repo at the ref
// given by the fragment of repo, branch or "master" (in that order)
func (gh Github) ArchiveURL(repo *url.URL, branch string) *url.URL {
	// NOTE: "master" as default one is a classic fallback, but
	// nowadays the primary branch is called "main" often. needs
	// to be tackled somewhen
	ref := utils.FirstNotEmpty(repo.Fragment, branch, "master")
	u := *repo
	u.Path = path.Join(u.Path, "archive", ref+".zip")
	return &u
}

// PluginArchiveURL returns the URL vopher fetches the github plugin p from:
// p.URL if it already points to an archive, ArchiveURL otherwise
func (gh Github) PluginArchiveURL(p *plugin.Plugin) *url.URL {
	if ok, _ := archive.IsSupportedArchive(path.Base(p.URL.Path)); ok {
		return p.URL
	}
	return gh.ArchiveURL(p.URL, p.Opts.Branch)
}

// IsCommit returns true if ref looks like a (short) git commit id. The
// archive of a commit never changes, unlike the one of a branch or tag.
func (gh Github) IsCommit(ref string) bool {
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
