package acquire

import (
	"os"
	"path/filepath"
	"slices"
	"testing"
)

// testdata/*.atom are feeds recorded from the hosters, reduced to 2 entries
// without author, content and summary
func TestParseFeed(t *testing.T) {

	tests := []struct {
		file     string
		entryID  func(id string) string
		expected []string
	}{
		{"github-commits.atom", commitID, []string{"3b753cf8c6a4dcde6edee8827d464ba9b8c4a6f0", "045d2de2d7b964259a5ba9d1af580ff87a80ffef"}},
		{"github-tags.atom", Github{}.TagName, []string{"pubsub/v2.8.0", "firestore/v1.28.0"}},
		{"forgejo-commits.atom", commitID, []string{"cfa2d58f4537aca6ca83e2de1a0d9f1491121264", "10c58a303f0d735e382b9ccd45ad4e13c240b309"}},
		{"forgejo-tags.atom", Forgejo{}.TagName, []string{"0.10.0", "0.9.0"}},
		{"gitea-tags.atom", Forgejo{}.TagName, []string{"gitea/v0.25.1", "gitea/v0.25.0"}},
		{"gitlab-commits.atom", commitID, []string{"a867a65a65d9d236164ed009fb534813d2e2c60b", "b1f8286df533c604f9f307d454c0fd8bef3723f2"}},
		{"gitlab-tags.atom", Gitlab{}.TagName, []string{"config/v2.1.0", "config/v2.0.0"}},
	}

	for i, test := range tests {
		f, err := os.Open(filepath.Join("testdata", test.file))
		if err != nil {
			t.Fatalf("%d: %v", i, err)
		}
		feed, err := parseFeed(f, test.entryID)
		f.Close()
		if err != nil {
			t.Fatalf("%d: %s: %v", i, test.file, err)
		}

		ids := []string{}
		for _, entry := range feed.Entry {
			ids = append(ids, entry.ID)
			if entry.Title == "" || entry.Updated == "" {
				t.Fatalf("%d: %s: entry %q without title or updated", i, test.file, entry.ID)
			}
		}
		t.Logf("test-%d: %s: %q", i, test.file, ids)
		if !slices.Equal(ids, test.expected) {
			t.Fatalf("%d: %s: expected %q, got %q", i, test.file, test.expected, ids)
		}
	}
}
