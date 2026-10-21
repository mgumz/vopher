package acquire

import (
	"net/url"
	"path"
	"strings"
)

// Forgejo groups the functions for Forgejo (codeberg.org) and Gitea, which
// share their URL layout
type Forgejo struct{}

// GetRepository extracts parts of `url`, see ownerRepository:
//
//	codeberg.org/username/reponame
//	codeberg.org/username/reponame/archive/main.zip
//	codeberg.org/username/reponame#v2.1
func (fj Forgejo) GetRepository(remote *url.URL) (name, head string) {
	return ownerRepository(remote)
}

// ArchiveURL returns codeberg.org/username/reponame/archive/<ref>.zip
func (fj Forgejo) ArchiveURL(repo *url.URL, ref string) url.URL {
	archiveURL := *repo
	archiveURL.Path = path.Join(archiveURL.Path, "archive", ref+".zip")
	return archiveURL
}

// CommitsFeedURL returns codeberg.org/username/reponame/atom/branch/<branch>
func (fj Forgejo) CommitsFeedURL(repo *url.URL, branch string) url.URL {
	feedURL := *repo
	feedURL.Path = path.Join(feedURL.Path, "atom", "branch", branch)
	return feedURL
}

// TagsFeedURL returns codeberg.org/username/reponame/tags.atom
func (fj Forgejo) TagsFeedURL(repo *url.URL) url.URL {
	feedURL := *repo
	feedURL.Path = path.Join(feedURL.Path, "tags") + ".atom"
	return feedURL
}

// TagName returns <tag> of "<release-id>: <repo>/releases/tag/<tag>"
func (fj Forgejo) TagName(id string) string {
	if _, tag, ok := strings.Cut(id, "/releases/tag/"); ok {
		return tag
	}
	return commitID(id)
}
