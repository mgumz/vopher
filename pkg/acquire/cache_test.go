package acquire

import (
	"net/url"
	"path/filepath"
	"strings"
	"testing"

	"github.com/mgumz/vopher/pkg/plugin"
)

func TestArchivePath(t *testing.T) {
	const cache = "/cache"
	tests := []struct {
		url, ext, want string
	}{
		{"https://github.com/tpope/vim-fugitive/archive/master.zip", ".zip", "/cache/github.com/tpope/vim-fugitive/archive/master.zip"},
		{"https://example.com/plugin.tar.gz", ".tar.gz", "/cache/example.com/plugin.tar.gz"},
		{"https://www.vim.org/scripts/download_script.php?src_id=123", ".zip", "/cache/www.vim.org/scripts/download_script.php_src_id%3D123.zip"},
		{"http://example.com:8080/../../etc/x.zip", ".zip", "/cache/example.com_8080/etc/x.zip"},
	}
	for _, tt := range tests {
		u, _ := url.Parse(tt.url)
		if got := ArchivePath(cache, u, tt.ext); got != filepath.FromSlash(tt.want) {
			t.Errorf("%s: got %q, want %q", tt.url, got, tt.want)
		}
	}
}

func TestIsArchiveOf(t *testing.T) {
	stem := filepath.FromSlash("/cache/www.vim.org/scripts/download_script.php_src_id%3D123")
	tests := map[string]bool{
		stem:                         true,
		stem + ".zip":                true,
		stem + ".tar.gz":             true,
		stem + ".zip-2026-10-20T05Z": false,
		stem + "4.zip":               false,
		stem[:len(stem)-1] + "4.zip": false,
	}
	for name, want := range tests {
		if got := IsArchiveOf(name, stem); got != want {
			t.Errorf("%q: got %v, want %v", name, got, want)
		}
	}
}

func TestPluginArchiveURLs(t *testing.T) {
	tests := []struct {
		url, branch, want string
	}{
		{"https://github.com/tpope/vim-fugitive", "", "https://github.com/tpope/vim-fugitive/archive/main.zip https://github.com/tpope/vim-fugitive/archive/master.zip"},
		{"https://github.com/tpope/vim-fugitive", "main", "https://github.com/tpope/vim-fugitive/archive/main.zip"},
		{"https://github.com/tpope/vim-fugitive#v2.1", "main", "https://github.com/tpope/vim-fugitive/archive/v2.1.zip#v2.1"},
		{"https://github.com/tpope/vim-fugitive/archive/v2.1.zip", "", "https://github.com/tpope/vim-fugitive/archive/v2.1.zip"},
	}
	for _, tt := range tests {
		u, _ := url.Parse(tt.url)
		p := &plugin.Plugin{URL: u, Opts: plugin.Opts{Branch: tt.branch}}
		urls := []string{}
		for _, u := range PluginArchiveURLs(Github{}, p) {
			urls = append(urls, u.String())
		}
		if got := strings.Join(urls, " "); got != tt.want {
			t.Errorf("%s: got %q, want %q", tt.url, got, tt.want)
		}
	}
}
