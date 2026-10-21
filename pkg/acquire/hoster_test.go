package acquire

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/mgumz/vopher/pkg/plugin"
)

func TestShortID(t *testing.T) {

	tests := []struct {
		id, expected string
	}{
		{"3b753cf8c6a4dcde6edee8827d464ba9b8c4a6f0", "3b753cf8c6"},
		{"v1.0.0", "v1.0.0"},
		{"v0.5.1-beta.2", "v0.5.1-beta.2"},
		{"firestore/v1.27.0", "firestore/v1.27.0"},
	}

	for i, test := range tests {
		id := shortID(test.id)
		t.Logf("test-%d: %q: %q", i, test.id, id)
		if id != test.expected {
			t.Fatalf("%d: expected %q, got %q", i, test.expected, id)
		}
	}
}

func TestHosterFor(t *testing.T) {

	tests := []struct {
		url, hoster string
		expected    Hoster
		fails       bool
	}{
		{"https://github.com/tpope/vim-fugitive", "", Github{}, false},
		{"https://GitHub.com/tpope/vim-fugitive", "", Github{}, false},
		{"https://codeberg.org/mfussenegger/nvim-dap", "", Forgejo{}, false},
		{"https://gitea.com/gitea/tea", "", Forgejo{}, false},
		{"https://gitlab.com/gitlab-org/editor-extensions/gitlab.vim", "", Gitlab{}, false},
		{"https://example.com/plugin.zip", "", nil, false},
		{"https://git.example.com/o/r", "github", Github{}, false},
		{"https://git.example.com/o/r", "Forgejo", nil, true},
		{"https://git.example.com/o/r", "nope", nil, true},
	}

	for i, test := range tests {
		u, _ := url.Parse(test.url)
		p := &plugin.Plugin{URL: u, Opts: plugin.Opts{Hoster: test.hoster}}
		hoster, err := HosterFor(p)
		t.Logf("test-%d: %q hoster=%q: %v %v", i, test.url, test.hoster, hoster, err)
		if (err != nil) != test.fails {
			t.Fatalf("%d: expected fail=%v, got %v", i, test.fails, err)
		}
		if hoster != test.expected {
			t.Fatalf("%d: expected %v, got %v", i, test.expected, hoster)
		}
	}
}

func TestProbeBranch(t *testing.T) {

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodHead && r.URL.Path == "/o/r/archive/master.zip" {
			return
		}
		http.NotFound(w, r)
	}))
	defer ts.Close()

	repo, _ := url.Parse(ts.URL + "/o/r")

	tests := []struct {
		branches []string
		expected string
		fails    bool
	}{
		{DefaultBranches, "master", false},
		{[]string{"master", "main"}, "master", false},
		{[]string{"main"}, "", true},
	}

	for i, test := range tests {
		branch, err := ProbeBranch(Github{}, repo, test.branches...)
		t.Logf("test-%d: %q: %q %v", i, test.branches, branch, err)
		if (err != nil) != test.fails || branch != test.expected {
			t.Fatalf("%d: expected %q fail=%v, got %q %v", i, test.expected, test.fails, branch, err)
		}
	}
}
