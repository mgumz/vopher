package main

import (
	"log"
	"os"
	"path"
	"path/filepath"
	"strings"
)

func act_update(plugins PluginList, dir string, force bool, ui JobUi) {

	ui.Start()

	for _, plugin := range plugins {

		plugin_folder := filepath.Join(dir, plugin.name)

		_, err := os.Stat(plugin_folder)
		if err == nil { // plugin_folder exists
			if !force {
				continue
			}
		}

		if !strings.HasSuffix(plugin.url.Path, ".zip") {
			switch plugin.url.Host {
			case "github.com":
				remote_zip := first_not_empty(plugin.url.Fragment, "master") + ".zip"
				plugin.url.Path = path.Join(plugin.url.Path, "archive", remote_zip)
			default:
				ext, err := httpdetect_ftype(plugin.url.String())
				if err != nil {
					log.Printf("error: %q: %s", plugin.url, err)
					continue
				}
				if ext != ".zip" {
					log.Printf("error: %q: not a zip", plugin.url)
					continue
				}
			}
		}

		ui.AddJob(plugin_folder)
		go acquire(plugin_folder, plugin.url.String(), plugin.strip_dir, ui)
	}
	ui.Wait()
	ui.Stop()
}
