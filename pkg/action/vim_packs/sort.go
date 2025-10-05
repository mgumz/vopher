package vim_packs

import (
	"cmp"
	"strings"

	"github.com/mgumz/vopher/pkg/plugin"
)

// helper to sort the packs by weight and name.
// idea is to use .add(), dependencies will be
// added "multiple" times (increasing the weight).
// if a dependency requires another one: it is increasing
// the weight of that one.
type sortedEntry struct {
	weight int
	p      *plugin.Plugin
}

type sortedEntries []sortedEntry

// satisfy slices.SortFunc. also, descending order
// for weight, regular "alphabetic" for plugin
// name.
func (a sortedEntry) cmp(b sortedEntry) int {

	if n := cmp.Compare(a.weight, b.weight); n != 0 {
		return n * -1
	}
	n := strings.Compare(a.p.Name, b.p.Name)
	return n
}

func (entries *sortedEntries) init(plugins plugin.List) {

	for _, p := range plugins {
		entries.add(p, 0)
		for _, depends := range p.Opts.DependsOn {
			if p, exists := plugins[depends]; exists {
				entries.add(p, 1)
			}
		}
	}
}

func (entries *sortedEntries) add(p *plugin.Plugin, weight int) {

	for i := range *entries {
		if (*entries)[i].p == p {
			(*entries)[i].weight += weight
			return
		}
	}

	*entries = append(*entries, sortedEntry{weight, p})
}
