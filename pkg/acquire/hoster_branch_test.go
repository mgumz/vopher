package acquire

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"slices"
	"testing"
)

func TestBranchCandidates(t *testing.T) {

	tests := []struct {
		branch, defaultBranch string
		expected              []string
	}{
		{"", "", []string{"main", "master"}},
		{"dev", "", []string{"dev", "main", "master"}},
		{"", "master", []string{"master", "main"}},
		{"dev", "trunk", []string{"dev", "trunk", "main", "master"}},
		{"main", "main", []string{"main", "master"}},
	}

	for i, test := range tests {
		candidates := branchCandidates(test.branch, test.defaultBranch)
		t.Logf("test-%d: %q %q: %q", i, test.branch, test.defaultBranch, candidates)
		if !slices.Equal(candidates, test.expected) {
			t.Fatalf("%d: expected %q, got %q", i, test.expected, candidates)
		}
	}
}

func TestPickBranch(t *testing.T) {

	tests := []struct {
		branches, candidates []string
		expected             string
	}{
		{[]string{"tests", "master"}, []string{"main", "master"}, "master"},
		{[]string{"tests", "master"}, []string{"tests", "main", "master"}, "tests"},
		{[]string{"feature-a", "feature-b"}, []string{"main", "master"}, "feature-a"},
		{[]string{}, []string{"main", "master"}, ""},
	}

	for i, test := range tests {
		branch := pickBranch(test.branches, test.candidates)
		t.Logf("test-%d: %q %q: %q", i, test.branches, test.candidates, branch)
		if branch != test.expected {
			t.Fatalf("%d: expected %q, got %q", i, test.expected, branch)
		}
	}
}

// testdata/ holds answers recorded from the hosters
func TestParseBranches(t *testing.T) {

	read := func(name string) []byte {
		body, err := os.ReadFile(filepath.Join("testdata", name))
		if err != nil {
			t.Fatal(err)
		}
		return body
	}

	branches := parseGithubBranches(read("github-branch-commits.html"))
	t.Logf("github: %q", branches)
	if !slices.Equal(branches, []string{"master"}) {
		t.Fatalf("github: expected [master], got %q", branches)
	}

	branches, defaultBranch, err := parseForgejoBranches(read("forgejo-branches.json"))
	t.Logf("forgejo: %q %q %v", branches, defaultBranch, err)
	if err != nil || !slices.Equal(branches, []string{"tests", "master"}) || defaultBranch != "master" {
		t.Fatalf("forgejo: expected [tests master] master, got %q %q %v", branches, defaultBranch, err)
	}

	branches, err = parseGitlabBranches(read("gitlab-refs.json"))
	t.Logf("gitlab: %q %v", branches, err)
	if err != nil || len(branches) != 4 || branches[3] != "main" {
		t.Fatalf("gitlab: expected 4 branches ending with main, got %q %v", branches, err)
	}
}

// gitea.com redirects load-branches-and-tags to its login page; BranchOf
// then asks the compare api for each candidate
func TestForgejoBranchOfCompare(t *testing.T) {

	const commit = "33bb12729e6de375813fb94fbcfbfa110d45c107"

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/o/r/commit/" + commit + "/load-branches-and-tags":
			http.Redirect(w, r, "/user/login", http.StatusSeeOther)
		case "/user/login":
			w.Write([]byte("<html>login</html>"))
		case "/api/v1/repos/o/r/compare/main..." + commit:
			w.Write([]byte(`{"total_commits":3,"commits":[]}`))
		case "/api/v1/repos/o/r/compare/master..." + commit:
			w.Write([]byte(`{"total_commits":0,"commits":[]}`))
		default:
			http.NotFound(w, r)
		}
	}))
	defer ts.Close()

	repo, _ := url.Parse(ts.URL + "/o/r")
	branch, err := Forgejo{}.BranchOf(repo, commit, "")
	t.Logf("%q %v", branch, err)
	if err != nil || branch != "master" {
		t.Fatalf("expected master, got %q %v", branch, err)
	}

	branch, err = Forgejo{}.BranchOf(repo, commit, "dev")
	if err != nil || branch != "master" {
		t.Fatalf("with branch=dev: expected master, got %q %v", branch, err)
	}
}
