package main

import (
	"log"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"
)

func act_update(plugins PluginList, dir string, force bool, ui JobUi) {

	ui.Start()

	for _, plugin := range plugins {

		plugin_folder := filepath.Join(dir, plugin.name)

		if _, err := os.Stat(plugin_folder); err == nil { // plugin_folder exists
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
		go acquire_and_postupdate(plugin_folder, plugin, ui)
	}

	ui.Wait()
	ui.Stop()
}

func acquire_and_postupdate(dir string, plugin Plugin, ui JobUi) {
	defer ui.JobDone(dir)

	var (
		err  error
		path string
		out  []byte
	)

	if err = acquire(dir, plugin.url.String(), plugin.strip_dir); err != nil {
		log.Printf("%s: %q", dir, err)
		return
	}

	//
	// handle the .postupdate hook
	//
	if plugin.postupdate == "" {
		return
	}

	path, err = expand_path(plugin.postupdate)
	if err != nil {
		log.Printf("%s: expanding .postupdate %q: %s", dir, plugin.postupdate, err)
		return
	}
	path = expand_path_environment(path, dir)

	// we won't check for existing executable, we just prepare the
	// right environment and then we launch it. if it fails, the OS will tell
	// us (missing permissions, file not found etc)
	//
	// we do not set .Dir because this will cause golang to search for .Path
	// underneath it. if the user wants to call a script inside .Dir (which
	// should be a rare case anyway) then she should use $VOPHER_DIR/cmd which
	// gets expanded, see above.
	cmd := exec.Cmd{
		Path: path,
		Env: append(os.Environ(),
			"VOPHER_NAME="+plugin.name,
			"VOPHER_DIR="+dir,
			"VOPHER_URL="+plugin.url.String()),
	}

	out, err = cmd.CombinedOutput()
	if err != nil {
		log.Printf("%s: error running .postupdate %q: %s", dir, path, err)
		log.Printf("%s: %s", dir, string(out))
		return
	}

	if len(out) == 0 {
		return
	}

	ui.Print(dir, string(out))
}
