package utils

import (
	"log"
	"os"
	"testing"
)

func TestIndexByteN(t *testing.T) {

	tests := []struct {
		idx  int
		text string
	}{
		{-1, ""},
		{-1, "/"},
		{-1, "1/2"},
		{3, "1/2/3"},
	}

	for i, test := range tests {
		idx := IndexByteN(test.text, '/', 2)
		t.Logf("test-%d: index of %d. '%c' in %q: %d",
			i, 2, '/', test.text, idx)
		if idx != test.idx {
			t.Fatalf("%d: expected %d, got %d", i, test.idx, idx)
		}
	}
}

func TestExpandVar(t *testing.T) {
	var (
		vopherDir = "YIPI"
		tests     = []struct {
			in, expected string
		}{
			{"foo", "foo"},
			{"foo bar", "foo bar"},
			{"foo $FOO", "foo BAR"},
			{"foo $VOPHER", "foo $VOPHER"},
			{"foo ${VOPHER}", "foo $VOPHER"},
			{"foo $VOPHER_DIR", "foo YIPI"},
		}
	)

	os.Clearenv()
	os.Setenv("FOO", "BAR")
	os.Setenv("FOO1", "WUFF")
	os.Setenv("EMPTY", "")

	for i := range tests {

		out := ExpandPathEnvironment(tests[i].in, vopherDir)

		t.Logf("%d: expand (VOPHER_DIR=%q, $ENV:%v) %q => %q",
			i, vopherDir, os.Environ(), tests[i].in, out)

		if out != tests[i].expected {
			log.Fatalf("%d: error expanding %q, got %q, expected %q\n%v",
				i, tests[i].in, out, tests[i].expected,
				os.Environ())
		}
	}
}

func TestParsePluginURL(t *testing.T) {

	tests := []struct {
		in, expected string
	}{
		{"github.com/tpope/vim-fugitive", "https://github.com/tpope/vim-fugitive"},
		{"codeberg.org/mfussenegger/nvim-dap", "https://codeberg.org/mfussenegger/nvim-dap"},
		{"gitea.com/gitea/tea", "https://gitea.com/gitea/tea"},
		{"gitlab.com/gitlab-org/editor-extensions/gitlab.vim", "https://gitlab.com/gitlab-org/editor-extensions/gitlab.vim"},
		{"GitLab.com/group/repo#v1.0", "https://GitLab.com/group/repo#v1.0"},
		{"https://codeberg.org/o/r", "https://codeberg.org/o/r"},
		{"example.com/plugin.zip", "example.com/plugin.zip"},
	}

	for i, test := range tests {
		u, err := ParsePluginURL(test.in)
		if err != nil {
			t.Fatalf("%d: %q: %v", i, test.in, err)
		}
		t.Logf("test-%d: %q: %q", i, test.in, u)
		if u.String() != test.expected {
			t.Fatalf("%d: expected %q, got %q", i, test.expected, u)
		}
	}
}
