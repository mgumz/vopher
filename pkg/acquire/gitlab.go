package acquire

import (
	"encoding/json"
	"fmt"
	"net/http"
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

// BranchOf asks the api
// gitlab.com/api/v4/projects/<group%2Freponame>/repository/commits/<commit>/refs?type=branch
func (gl Gitlab) BranchOf(repo *url.URL, commit, branch string) (string, error) {
	project := url.PathEscape(strings.TrimPrefix(repo.Path, "/"))
	u := repo.Scheme + "://" + repo.Host + "/api/v4/projects/" + project +
		"/repository/commits/" + url.PathEscape(commit) + "/refs?type=branch&per_page=100"
	body, status, err := fetchBody(u)
	if err != nil {
		return "", err
	}
	if status != http.StatusOK {
		return "", fmt.Errorf("%q: %d", u, status)
	}
	branches, err := parseGitlabBranches(body)
	if err != nil {
		return "", fmt.Errorf("%q: %w", u, err)
	}
	return pickBranch(branches, branchCandidates(branch, "")), nil
}

func parseGitlabBranches(body []byte) ([]string, error) {
	var refs []struct {
		Name string `json:"name"`
	}
	if err := json.Unmarshal(body, &refs); err != nil {
		return nil, err
	}
	branches := []string{}
	for _, r := range refs {
		branches = append(branches, r.Name)
	}
	return branches, nil
}
