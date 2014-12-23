package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
)

func act_prune(plugins PluginList, base string, force bool, all bool) {

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
			if _, is_plugin_zip := plugins[name[:len(name)-4]]; is_plugin_zip {
				continue
			}
		}

		if _, is_plugin := plugins[name]; !is_plugin {
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
