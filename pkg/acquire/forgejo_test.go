package acquire

import (
	"net/url"
	"testing"
)

func TestForgejoURLs(t *testing.T) {

	repo, _ := url.Parse("https://codeberg.org/mfussenegger/nvim-dap")
	fj := Forgejo{}

	tests := []struct {
		u        url.URL
		expected string
	}{
		{fj.ArchiveURL(repo, "master"), "https://codeberg.org/mfussenegger/nvim-dap/archive/master.zip"},
		{fj.CommitsFeedURL(repo, "master"), "https://codeberg.org/mfussenegger/nvim-dap/atom/branch/master"},
		{fj.TagsFeedURL(repo), "https://codeberg.org/mfussenegger/nvim-dap/tags.atom"},
	}

	for i, test := range tests {
		u := test.u.String()
		t.Logf("test-%d: %q", i, u)
		if u != test.expected {
			t.Fatalf("%d: expected %q, got %q", i, test.expected, u)
		}
	}
}

func TestForgejoFeedIDs(t *testing.T) {

	tests := []struct {
		entryID      func(id string) string
		id, expected string
	}{
		{commitID, "cfa2d58f4537aca6ca83e2de1a0d9f1491121264", "cfa2d58f4537aca6ca83e2de1a0d9f1491121264"},
		{Forgejo{}.TagName, "3157808: https://codeberg.org/mfussenegger/nvim-dap/releases/tag/0.10.0", "0.10.0"},
		{Forgejo{}.TagName, "868261: https://gitea.com/gitea/go-sdk/releases/tag/gitea/v0.25.1", "gitea/v0.25.1"},
	}

	for i, test := range tests {
		id := test.entryID(test.id)
		t.Logf("test-%d: %q: %q", i, test.id, id)
		if id != test.expected {
			t.Fatalf("%d: expected %q, got %q", i, test.expected, id)
		}
	}
}
