package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"
)

type actUpdateOptions struct {
	dir     string // dir to extract the plugins to
	force   bool   // enforce acquire, even plugin exists
	dry_run bool   // don't extract, just show the files
	sha1    string // checksum to check the downloaded file against
}

func act_update(plugins PluginList, ui JobUi, opts *actUpdateOptions) {

	ui.Start()

	for _, plugin := range plugins {

		plugin_folder := filepath.Join(opts.dir, plugin.name)

		if _, err := os.Stat(plugin_folder); err == nil { // plugin_folder exists
			if !opts.force {
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
		go acquire_and_postupdate(plugin_folder, plugin.sha1, opts.dry_run, plugin, ui)
	}

	ui.Wait()
	ui.Stop()
}

func acquire_and_postupdate(dir, sha1 string, dry_run bool, plugin *Plugin, ui JobUi) {

	defer ui.JobDone(dir)

	var (
		err  error
		path string
		out  []byte

		acquire_f func(string, string, int, string) error = acquire
	)

	if dry_run {
		acquire_f = dry_acquire
	}
	if err = acquire_f(dir, plugin.url.String(), plugin.strip_dir, sha1); err != nil {
		log.Printf("%s: %v", dir, err)
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

	if dry_run {
		fmt.Printf("# postupdate: %q (env: %v)\n", cmd.Path, cmd.Env)
		return
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
