package acquire

import (
	"net/url"
	"testing"
)

func TestGitlabGetRepository(t *testing.T) {

	tests := []struct {
		url, name, head string
	}{
		{"https://gitlab.com/group/repo", "/group/repo", "master"},
		{"https://gitlab.com/group/repo/", "/group/repo", "master"},
		{"https://gitlab.com/group/sub/repo", "/group/sub/repo", "master"},
		{"https://gitlab.com/group/sub/repo#v2.1", "/group/sub/repo", "v2.1"},
		{"https://gitlab.com/group/repo/-/archive/main.zip", "/group/repo", "main"},
		{"https://gitlab.com/group/repo/-/archive/config/v2.1.0.zip", "/group/repo", "config/v2.1.0"},
		{"https://gitlab.com/group/repo/-/archive/main/repo-main.zip", "/group/repo", "main"},
		{"https://gitlab.com/group/repo/-/archive/config/v2.1.0/repo-config-v2.1.0.zip", "/group/repo", "config/v2.1.0"},
	}

	for i, test := range tests {
		u, _ := url.Parse(test.url)
		name, head := Gitlab{}.GetRepository(u)
		t.Logf("test-%d: %q: %q %q", i, test.url, name, head)
		if name != test.name || head != test.head {
			t.Fatalf("%d: expected %q %q, got %q %q", i, test.name, test.head, name, head)
		}
	}
}

func TestGitlabURLs(t *testing.T) {

	repo, _ := url.Parse("https://gitlab.com/gitlab-org/editor-extensions/gitlab.vim")
	gl := Gitlab{}

	tests := []struct {
		u        url.URL
		expected string
	}{
		{gl.ArchiveURL(repo, "main"), "https://gitlab.com/gitlab-org/editor-extensions/gitlab.vim/-/archive/main.zip"},
		{gl.CommitsFeedURL(repo, "main"), "https://gitlab.com/gitlab-org/editor-extensions/gitlab.vim/-/commits/main?format=atom"},
		{gl.TagsFeedURL(repo), "https://gitlab.com/gitlab-org/editor-extensions/gitlab.vim/-/tags?format=atom"},
	}

	for i, test := range tests {
		u := test.u.String()
		t.Logf("test-%d: %q", i, u)
		if u != test.expected {
			t.Fatalf("%d: expected %q, got %q", i, test.expected, u)
		}
	}
}

func TestGitlabFeedIDs(t *testing.T) {

	tests := []struct {
		entryID      func(id string) string
		id, expected string
	}{
		{commitID, "https://gitlab.com/gitlab-org/gitlab-shell/-/commit/20106983a3e5ce9f32946673229cbda3bb249ef8", "20106983a3e5ce9f32946673229cbda3bb249ef8"},
		{Gitlab{}.TagName, "https://gitlab.com/gitlab-org/gitlab-shell/-/tags/v14.59.0", "v14.59.0"},
		{Gitlab{}.TagName, "https://gitlab.com/gitlab-org/api/client-go/-/tags/config%2Fv2.1.0", "config/v2.1.0"},
	}

	for i, test := range tests {
		id := test.entryID(test.id)
		t.Logf("test-%d: %q: %q", i, test.id, id)
		if id != test.expected {
			t.Fatalf("%d: expected %q, got %q", i, test.expected, id)
		}
	}
}
