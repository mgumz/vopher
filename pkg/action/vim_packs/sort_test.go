package vim_packs

import (
	"testing"

	"github.com/mgumz/vopher/pkg/plugin"
)

func Test_SortPacks(t *testing.T) {

	plugins := plugin.List{
		"a": &plugin.Plugin{
			Name: "a",
			Opts: plugin.Opts{
				DependsOn: []string{"b"},
			},
		},
		"b": &plugin.Plugin{Name: "b"},
		"c": &plugin.Plugin{Name: "a",
			Opts: plugin.Opts{
				DependsOn: []string{"a", "b"},
			}},
	}

	ts := newTarjanSolver(plugins)
	packs, cycles, err := ts.solve()

	if err != nil {
		t.Fatalf("expected err being nil, actual it is %v", err)
	}

	if len(cycles) > 0 {
		t.Fatalf("expected no cycles, actual %v", cycles)
	}

	packs.sort()

	// a -> b
	// c -> a
	// c -> b
	// b
	expectedOrder := []string{"b", "a", "c"}

	t.Logf("%#v", plugins)
	t.Logf("%#v", packs)

	for i := range packs {
		t.Logf("%d %s", i, packs[i].Name)
		if expectedOrder[i] != packs[i].Name {
			t.Fatalf("%d: expected %q, actual %q",
				i, expectedOrder[i], packs[i].Name)
		}
	}

}
