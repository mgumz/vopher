package acquire

import (
	"net/url"
	"path"
	"strings"

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

// ArchiveURL returns github.com/username/reponame/archive/<ref>.zip
func (gh Github) ArchiveURL(repo *url.URL, ref string) url.URL {
	archiveURL := *repo
	archiveURL.Path = path.Join(archiveURL.Path, "archive", ref+".zip")
	return archiveURL
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
