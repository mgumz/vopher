package action

import (
	"fmt"
	"log"
	"os"
	"sort"

	"github.com/mgumz/vopher/pkg/plugin"
)

func Status(plugins plugin.List, base string) {

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

	entries := make(map[string]*plugin.DirEntry)

	for i := range dirEntries {

		if !dirEntries[i].IsDir() {
			continue
		}

		name := dirEntries[i].Name()
		isPlugin := plugins.Exists(name)
		entry := &plugin.DirEntry{
			Name:     name,
			Exists:   1,
			IsPlugin: boolAsInt(isPlugin),
		}

		entries[name] = entry
	}

	for name := range plugins {
		if _, exists := entries[name]; !exists {
			entries[name] = &plugin.DirEntry{
				Name:     name,
				Exists:   0,
				IsPlugin: 1,
			}
		}
	}

	ordered := plugin.DirEntryByName{}
	for _, entry := range entries {
		ordered = append(ordered, entry)
	}
	sort.Sort(ordered)

	state := " vm " // v-vopher handled; m-missing
	for i := range ordered {
		fmt.Printf("%c%c %s\n",
			state[ordered[i].IsPlugin],
			state[2+ordered[i].Exists],
			ordered[i].Name)
	}
}
