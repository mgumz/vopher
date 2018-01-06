package main

import (
	"log"
	"os"
	"path/filepath"
)

func actClean(plugins PluginList, dir string, force bool) {

	if !force {
		log.Println("'clean' needs -force flag")
		return
	}

	var prefix, suffix string

	for _, plugin := range plugins {
		folder := filepath.Join(dir, plugin.name)
		prefix = ""
		suffix = "ok"
		_, err := os.Stat(folder)
		if err == nil { // plugin_folder exists
			err = os.RemoveAll(folder)
			if err != nil {
				prefix = "error:"
				suffix = err.Error()
			}
		} else {
			prefix = "info:"
			suffix = "does not exist"
		}
		log.Println("'clean'", prefix, folder, suffix)
	}
}
