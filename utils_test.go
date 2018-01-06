package main

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
		idx := indexByteN(test.text, '/', 2)
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

		out := expandPathEnvironment(tests[i].in, vopherDir)

		t.Logf("%d: expand (VOPHER_DIR=%q, $ENV:%v) %q => %q",
			i, vopherDir, os.Environ(), tests[i].in, out)

		if out != tests[i].expected {
			log.Fatalf("%d: error expanding %q, got %q, expected %q\n%v",
				i, tests[i].in, out, tests[i].expected,
				os.Environ())
		}
	}
}
