package plugin

import (
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

func (plugins List) Exists(name string) bool {
	_, exists := plugins[name]
	return exists
}
