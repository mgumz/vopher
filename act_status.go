package main

import (
	"fmt"
	"log"
	"os"
	"sort"
)

func actStatus(plugins PluginList, base string) {

	dir, err := os.Open(base)
	if err != nil {
		log.Println(err)
		return
	}
	defer dir.Close()

	dirEntries, err := dir.Readdir(-1)
	if err != nil {
		log.Println(err)
		return
	}

	boolAsInt := func(b bool) int {
		if b {
			return 1
		}
		return 0
	}

	entries := make(map[string]*PluginDirEntry)

	for i := range dirEntries {

		if !dirEntries[i].IsDir() {
			continue
		}

		name := dirEntries[i].Name()
		isPlugin := plugins.exists(name)
		entry := &PluginDirEntry{
			name:     name,
			exists:   1,
			isPlugin: boolAsInt(isPlugin),
		}

		entries[name] = entry
	}

	for name := range plugins {
		if _, exists := entries[name]; !exists {
			entries[name] = &PluginDirEntry{
				name:     name,
				exists:   0,
				isPlugin: 1,
			}
		}
	}

	ordered := PluginDirEntryByName{}
	for _, entry := range entries {
		ordered = append(ordered, entry)
	}
	sort.Sort(ordered)

	state := " vm " // v-vopher handled; m-missing
	for i := range ordered {
		fmt.Printf("%c%c %s\n",
			state[ordered[i].isPlugin],
			state[2+ordered[i].exists],
			ordered[i].name)
	}
}
