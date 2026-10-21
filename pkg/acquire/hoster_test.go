package acquire

import (
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
		{"https://example.com/plugin.zip", "", nil, false},
		{"https://git.example.com/o/r", "github", Github{}, false},
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
