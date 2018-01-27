package main

import (
	"sort"
	"strings"
)

type PluginList map[string]*Plugin

type PluginParser func(PluginList, string) error

func (plugins PluginList) filter(filter stringList) PluginList {

	if len(filter) == 0 {
		return plugins
	}

	filtered := make(PluginList)
	for k, v := range plugins {
		for i := range filter {
			if strings.Contains(k, filter[i]) {
				filtered[k] = v
			}
		}
	}
	return filtered
}

func (plugins PluginList) sortedIDs() []string {
	ids := []string{}
	for id := range plugins {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	return ids
}

func (plugins PluginList) exists(name string) bool {
	_, exists := plugins[name]
	return exists
}
