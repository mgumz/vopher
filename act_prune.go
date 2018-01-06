package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
)

func actPrune(plugins PluginList, base string, force, all bool) {

	dir, err := os.Open(base)
	if err != nil {
		log.Println(err)
		return
	}
	defer dir.Close()

	entries, err := dir.Readdir(-1)
	if err != nil {
		log.Println(err)
		return
	}

	for i := range entries {

		name := filepath.Base(entries[i].Name())

		// spare plugin.zip from pruning
		if !all && strings.HasSuffix(name, ".zip") {
			if _, isPlugin := plugins[name[:len(name)-4]]; isPlugin {
				continue
			}
		}

		if _, isPlugin := plugins[name]; !isPlugin {
			suffix := "dry-run."
			name = filepath.Join(base, name)
			fmt.Printf("prune %q: ", name)
			if force {
				suffix = "done."
				if err := os.RemoveAll(name); err != nil {
					suffix = err.Error()
				}
			}
			fmt.Println(suffix)
		}
	}
}
