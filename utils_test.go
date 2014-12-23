package main

import "testing"

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
		idx := index_byte_n(test.text, '/', 2)
		t.Logf("test-%d: index of %d. '%c' in %q: %d",
			i, 2, '/', test.text, idx)
		if idx != test.idx {
			t.Fatalf("%d: expected %d, got %d", i, test.idx, idx)
		}
	}
}
