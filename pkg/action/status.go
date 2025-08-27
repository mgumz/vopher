package action

import (
	"fmt"
	"io/fs"
	"log"
	"os"
	"sort"

	"github.com/mgumz/vopher/pkg/plugin"
)

// Status prints the status of all given plugins
func Status(plugins plugin.List, base string) {

	boolAsInt := func(b bool) int {
		if b {
			return 1
		}
		return 0
	}

	entries := make(map[string]*plugin.DirEntry)

	dfs := os.DirFS(base)

	// step-1: check for the presence of plugin-folders. note:
	// we have to go way down into `base` because the plugins
	// could be stored in subsubfolders. eg, $base/common/opt/fugitive
	// so it is insufficient to collect the direntries of $base
	// to deduce if a plugin exists, is missing or alike.
	for p := range plugins {
		fi, err := fs.Stat(dfs, p)
		entry := &plugin.DirEntry{
			Name:     p,
			Exists:   boolAsInt(fi != nil),
			IsPlugin: 1,
		}

		if err == fs.ErrNotExist {
			entry.Exists = 0
		}
		if fi != nil && !fi.IsDir() { // non directories are not plugins
			entry.IsPlugin = 0
		}
		entries[entry.Name] = entry
	}

	// step-2: massage the first-level entries of the
	// base directory in
	dirEntries, err := fs.ReadDir(dfs, ".")

	if err != nil {
		log.Println(err)
		return
	}

	for i := range dirEntries {

		// skip files
		if !dirEntries[i].IsDir() {
			continue
		}

		name := dirEntries[i].Name()

		// skip already checked in step-1 entries
		if _, already := entries[name]; already {
			continue
		}

		entry := &plugin.DirEntry{
			Name:     name,
			Exists:   1,
			IsPlugin: boolAsInt(plugins.Exists(name)),
		}

		entries[name] = entry
	}

	// sort output
	ordered := make(plugin.DirEntryByName, 0, len(entries))
	for _, entry := range entries {
		ordered = append(ordered, entry)
	}
	sort.Sort(ordered)

	// print
	state := " vm " // v-vopher handled; m-missing
	for i := range ordered {
		fmt.Printf("%c%c %s\n",
			state[ordered[i].IsPlugin],
			state[2+ordered[i].Exists],
			ordered[i].Name)
	}
}
