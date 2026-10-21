package acquire

import (
	"net/url"
	"testing"
)

func TestGithubFeedIDs(t *testing.T) {

	tests := []struct {
		entryID      func(id string) string
		id, expected string
	}{
		{commitID, "tag:github.com,2008:Grit::Commit/3b753cf8c6a4dcde6edee8827d464ba9b8c4a6f0", "3b753cf8c6a4dcde6edee8827d464ba9b8c4a6f0"},
		{Github{}.TagName, "tag:github.com,2008:Repository/331603/v3.7", "v3.7"},
		{Github{}.TagName, "tag:github.com,2008:Repository/19608522/firestore/v1.27.0", "firestore/v1.27.0"},
	}

	for i, test := range tests {
		id := test.entryID(test.id)
		t.Logf("test-%d: %q: %q", i, test.id, id)
		if id != test.expected {
			t.Fatalf("%d: expected %q, got %q", i, test.expected, id)
		}
	}
}

func TestGithubURLs(t *testing.T) {

	repo, _ := url.Parse("https://github.com/tpope/vim-fugitive")
	gh := Github{}

	tests := []struct {
		u        url.URL
		expected string
	}{
		{gh.ArchiveURL(repo, "master"), "https://github.com/tpope/vim-fugitive/archive/master.zip"},
		{gh.ArchiveURL(repo, "firestore/v1.27.0"), "https://github.com/tpope/vim-fugitive/archive/firestore/v1.27.0.zip"},
		{gh.CommitsFeedURL(repo, "main"), "https://github.com/tpope/vim-fugitive/commits/main.atom"},
		{gh.TagsFeedURL(repo), "https://github.com/tpope/vim-fugitive/tags.atom"},
	}

	for i, test := range tests {
		u := test.u.String()
		t.Logf("test-%d: %q", i, u)
		if u != test.expected {
			t.Fatalf("%d: expected %q, got %q", i, test.expected, u)
		}
	}
}
