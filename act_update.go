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

	var err error

	for _, plugin := range plugins {

		plugin_folder := filepath.Join(opts.dir, plugin.name)

		if _, err = os.Stat(plugin_folder); err == nil { // plugin_folder exists
			if !opts.force {
				continue
			}
		}

		archive_name := filepath.Base(plugin.url.Path)
		plugin.ext = filepath.Ext(archive_name)

		// apply heuristics aka ""guess""
		if plugin.ext == "" {
			switch plugin.url.Host {
			case "github.com":
				remote_zip := first_not_empty(plugin.url.Fragment, "master") + ".zip"
				plugin.url.Path = path.Join(plugin.url.Path, "archive", remote_zip)
				archive_name = filepath.Base(remote_zip)
				plugin.ext = ".zip"
				plugin.archive = &ZipArchive{}
			default:
				plugin.ext, err = httpdetect_ftype(plugin.url.String())
				if err != nil {
					log.Printf("error: %q: %s", plugin.url, err)
					continue
				}
				archive_name += plugin.ext
			}
		}

		if ok, suffix_len := IsSupportedArchive(archive_name); ok {
			plugin.ext = archive_name[len(archive_name)-suffix_len:]
		}

		if plugin.archive == nil {
			plugin.archive, err = GuessPluginArchive(archive_name)
			if err != nil {
				log.Printf("error: %q: not supported archive format", plugin.url)
				continue
			}
		}

		ui.AddJob(plugin_folder)
		go acquire_and_postupdate(plugin_folder, opts.dry_run, plugin, ui)
	}

	ui.Wait()
	ui.Stop()
}

func acquire_and_postupdate(dir string, dry_run bool, plugin *Plugin, ui JobUi) {

	defer ui.JobDone(dir)

	var (
		err     error
		path    string
		out     []byte
		entries []string

		url       = plugin.url.String()
		strip_dir = plugin.opts.strip_dir
		sha1      = plugin.opts.sha1
	)

	if dry_run {
		entries, err = dry_acquire(dir, url, plugin.archive, strip_dir, sha1)
	} else {
		err = acquire(dir, plugin.ext, url, plugin.archive, strip_dir, sha1)
	}
	if err != nil {
		log.Printf("%s: %v", dir, err)
		return
	}
	if dry_run {
		ui.Print(dir, strings.Join(entries, "\n"))
	}

	//
	// handle the .postupdate hook
	//
	if plugin.opts.postupdate == "" {
		return
	}

	path, err = expand_path(plugin.opts.postupdate)
	if err != nil {
		log.Printf("%s: expanding .postupdate %q: %s", dir, plugin.opts.postupdate, err)
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
			"VOPHER_ARCHIVE="+plugin.name+plugin.ext,
			"VOPHER_DIR="+dir,
			"VOPHER_URL="+url),
	}

	if dry_run {
		ui.Print(dir, fmt.Sprintf("# postupdate: %q (env: %v)\n", cmd.Path, cmd.Env))
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
