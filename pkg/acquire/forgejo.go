package acquire

import (
	"encoding/json"
	"net/http"
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

// BranchOf asks codeberg.org/username/reponame/commit/<commit>/load-branches-and-tags.
// gitea.com answers it only after a login; then it asks the api for each
// candidate whether it holds commit.
func (fj Forgejo) BranchOf(repo *url.URL, commit, branch string) (string, error) {
	u := *repo
	u.Path = path.Join(u.Path, "commit", commit, "load-branches-and-tags")
	body, status, err := fetchBody(u.String())
	if err != nil {
		return "", err
	}
	if status == http.StatusOK {
		if branches, defaultBranch, err := parseForgejoBranches(body); err == nil {
			return pickBranch(branches, branchCandidates(branch, defaultBranch)), nil
		}
	}
	return fj.compareBranch(repo, commit, branchCandidates(branch, ""))
}

func parseForgejoBranches(body []byte) (branches []string, defaultBranch string, err error) {
	var answer struct {
		Branches []struct {
			Name string `json:"name"`
		} `json:"branches"`
		DefaultBranch string `json:"default_branch"`
	}
	if err = json.Unmarshal(body, &answer); err != nil {
		return nil, "", err
	}
	for _, b := range answer.Branches {
		branches = append(branches, b.Name)
	}
	return branches, answer.DefaultBranch, nil
}

// compareBranch returns the first of candidates holding commit:
// /api/v1/repos/<owner>/<repo>/compare/<candidate>...<commit> has no commits
// for it
func (fj Forgejo) compareBranch(repo *url.URL, commit string, candidates []string) (string, error) {
	for _, c := range candidates {
		u := *repo
		u.Path = path.Join("/api/v1/repos", repo.Path, "compare", c+"..."+commit)
		body, status, err := fetchBody(u.String())
		if err != nil {
			return "", err
		}
		if status != http.StatusOK {
			continue
		}
		var answer struct {
			TotalCommits *int `json:"total_commits"`
		}
		if json.Unmarshal(body, &answer) == nil && answer.TotalCommits != nil && *answer.TotalCommits == 0 {
			return c, nil
		}
	}
	return "", nil
}
