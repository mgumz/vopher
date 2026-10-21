package acquire

import "testing"

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
