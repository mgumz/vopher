package main

import (
	"fmt"
	"log"
	"os"
	"sort"
)

type PluginDirEntry struct {
	name      string
	exists    int
	is_plugin int
	is_dir    int
}

type PluginDirEntryByName []*PluginDirEntry

func (a PluginDirEntryByName) Len() int           { return len(a) }
func (a PluginDirEntryByName) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a PluginDirEntryByName) Less(i, j int) bool { return a[i].name < a[j].name }

func act_status(plugins PluginList, base string) {

	dir, err := os.Open(base)
	if err != nil {
		log.Println(err)
		return
	}
	defer dir.Close()

	dir_entries, err := dir.Readdir(-1)
	if err != nil {
		log.Println(err)
		return
	}

	entries := make(map[string]*PluginDirEntry)

	for i := range dir_entries {

		if !dir_entries[i].IsDir() {
			continue
		}

		_, is_plugin := plugins[dir_entries[i].Name()]

		entry := PluginDirEntry{
			name:      dir_entries[i].Name(),
			exists:    1,
			is_plugin: bool_as_int(is_plugin),
		}

		entries[entry.name] = &entry
	}

	for name, _ := range plugins {
		if _, exists := entries[name]; !exists {
			entries[name] = &PluginDirEntry{
				name:      name,
				exists:    0,
				is_plugin: 1,
			}
		}
	}

	ordered := make(PluginDirEntryByName, len(entries))
	i := 0
	for _, entry := range entries {
		ordered[i] = entry
		i++
	}
	sort.Sort(ordered)

	const state = " vm "
	for i := range ordered {
		fmt.Printf("%c%c %s\n", state[ordered[i].is_plugin],
			state[2+ordered[i].exists], ordered[i].name)
	}

}

func bool_as_int(b bool) int {
	if b {
		return 1
	}
	return 0
}
