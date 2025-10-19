package plugin

import (
	"path"
	"sort"
	"strings"

	"github.com/mgumz/vopher/pkg/utils"
)

type List map[string]*Plugin

func (plugins List) Filter(filter utils.StringList) List {

	if len(filter) == 0 {
		return plugins
	}

	filtered := make(List)
	for k, v := range plugins {
		for i := range filter {
			if strings.Contains(k, filter[i]) {
				filtered[k] = v
			}
		}
	}
	return filtered
}

func (plugins List) SortedIDs() []string {
	ids := []string{}
	for id := range plugins {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	return ids
}

type idLine struct {
	id   string
	line int
}
type byLineNumber []idLine

func (a byLineNumber) Len() int           { return len(a) }
func (a byLineNumber) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a byLineNumber) Less(i, j int) bool { return a[i].line < a[j].line }

func (plugins List) SortByLineNumber() []string {

	bln := byLineNumber{}
	for _, p := range plugins {
		bln = append(bln, idLine{p.Name, p.ln})
	}

	sort.Sort(bln)

	ids := []string{}
	for _, p := range bln {
		ids = append(ids, p.id)
	}

	return ids
}

// Exists(p) checks if a plugin named "p" exists in the plugin list using
// path-based matching. The Key Idea: Instead of exact name matching, it
// compares the last path segment of each plugin against p. So whether a
// plugin is registered as wuzz, foo/wuzz, or foo/bar/wuzz, checking
// Exists("wuzz") will find it. Why: This simplifies dependency
// specifications — users only need to reference the final path segment (like
// depends-on=wuzz), which aligns with how vim/nvim's packadd! command works.
func (plugins List) Exists(p string) bool {
	for name, _ := range plugins {
		if p == name || path.Base(name) == p {
			return true
		}
	}
	return false
}
