package acquire

import (
	"net/url"
	"path"
	"strings"
)

// Gitlab groups the functions for GitLab (gitlab.com)
type Gitlab struct{}

// GetRepository extracts parts of `url`. Repositories may sit in nested
// groups, so the repo-name is the repo.Path up to "/-/":
//
//	gitlab.com/group/reponame
//	gitlab.com/group/subgroup/reponame
//	gitlab.com/group/reponame/-/archive/main.zip
//	gitlab.com/group/reponame/-/archive/main/reponame-main.zip
//	gitlab.com/group/reponame#v2.1
func (gl Gitlab) GetRepository(remote *url.URL) (name, head string) {
	name, rest, _ := strings.Cut(remote.Path, "/-/")
	name = strings.TrimSuffix(name, "/")

	head = "master"
	if remote.Fragment != "" {
		head = remote.Fragment
	} else if ref, ok := strings.CutPrefix(rest, "archive/"); ok && strings.HasSuffix(ref, ".zip") {
		head = ref[:len(ref)-4]
		// long form: <ref>/<reponame>-<ref with "/" as "-">
		if dir, file := path.Split(head); dir != "" {
			dir = strings.TrimSuffix(dir, "/")
			if file == path.Base(name)+"-"+strings.ReplaceAll(dir, "/", "-") {
				head = dir
			}
		}
	}

	return name, head
}

// ArchiveURL returns gitlab.com/group/reponame/-/archive/<ref>.zip
func (gl Gitlab) ArchiveURL(repo *url.URL, ref string) url.URL {
	archiveURL := *repo
	archiveURL.Path = path.Join(archiveURL.Path, "-", "archive", ref+".zip")
	return archiveURL
}

// CommitsFeedURL returns gitlab.com/group/reponame/-/commits/<branch>?format=atom
func (gl Gitlab) CommitsFeedURL(repo *url.URL, branch string) url.URL {
	return gl.feedURL(repo, "commits", branch)
}

// TagsFeedURL returns gitlab.com/group/reponame/-/tags?format=atom
func (gl Gitlab) TagsFeedURL(repo *url.URL) url.URL {
	return gl.feedURL(repo, "tags")
}

func (gl Gitlab) feedURL(repo *url.URL, parts ...string) url.URL {
	feedURL := *repo
	feedURL.Path = path.Join(feedURL.Path, "-", path.Join(parts...))
	feedURL.RawQuery = "format=atom"
	return feedURL
}

// TagName returns <tag> of "<repo>/-/tags/<tag>". GitLab escapes "/" in the
// tag name as "%2F".
func (gl Gitlab) TagName(id string) string {
	_, tag, ok := strings.Cut(id, "/-/tags/")
	if !ok {
		return commitID(id)
	}
	if unescaped, err := url.PathUnescape(tag); err == nil {
		return unescaped
	}
	return tag
}
