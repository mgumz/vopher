package plugin

import "testing"

func Test_ListExists(t *testing.T) {

	packs := List{
		"a": &Plugin{
			Name: "a",
		},
		"foo/bar/wuzz": &Plugin{
			Name: "foo/bar/wuzz",
		},
	}

	fixtures := []struct {
		toCheck  string
		expected bool
	}{
		{"a", true},
		{"does-not-exist", false},
		{"wuzz", true},
		{"bar/wuzz", false},
		{"foo/bar", false},
		{"/foo/bar", false},
	}

	t.Logf("packs: %#v\n", packs)
	for i, f := range fixtures {
		t.Logf("%d %q: expect %v", i, f.toCheck, f.expected)
		if exists := packs.Exists(f.toCheck); exists != f.expected {
			t.Fatalf("failure test-case %d: checked for %q, got %v, expected %v\n",
				i, f.toCheck, exists, f.expected)
		}
	}

}
