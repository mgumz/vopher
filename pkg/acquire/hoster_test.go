package acquire

import "testing"

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
