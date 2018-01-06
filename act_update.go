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

type ActUpdateOpts struct {
	dir    string // dir to extract the plugins to
	force  bool   // enforce acquire, even plugin exists
	dryRun bool   // don't extract, just show the files
	sha1   string // checksum to check the downloaded file against
}

func actUpdate(plugins PluginList, ui JobUI, opts *ActUpdateOpts) {

	ui.Start()

	var err error

	for _, plugin := range plugins {

		pluginFolder := filepath.Join(opts.dir, plugin.name)

		if _, err = os.Stat(pluginFolder); err == nil { // plugin_folder exists
			if !opts.force {
				continue
			}
		}

		archiveName := filepath.Base(plugin.url.Path)

		// apply heuristics aka ""guess""
		if isArchive, _ := isSupportedArchive(archiveName); !isArchive {
			switch plugin.url.Host {
			case "github.com":
				remoteZip := firstNotEmpty(plugin.url.Fragment, "master") + ".zip"
				plugin.url.Path = path.Join(plugin.url.Path, "archive", remoteZip)
				archiveName = filepath.Base(remoteZip)
				plugin.ext, plugin.archive = ".zip", &ZipArchive{}
			default:
				plugin.ext, err = httpdetectFtype(plugin.url.String())
				if err != nil {
					log.Printf("error: %q: %s", plugin.url, err)
					continue
				}
				archiveName += plugin.ext
			}
		}

		if ok, suffixLen := isSupportedArchive(archiveName); ok {
			plugin.ext = archiveName[len(archiveName)-suffixLen:]
		}

		if plugin.archive == nil {
			plugin.archive, err = guessPluginArchive(archiveName)
			if err != nil {
				log.Printf("error: %q: not supported archive format", plugin.url)
				continue
			}
		}

		ui.AddJob(pluginFolder)
		go acquireAndPostupdate(pluginFolder, opts.dryRun, plugin, ui)
	}

	ui.Wait()
	ui.Stop()
}

func acquireAndPostupdate(dir string, dryRun bool, plugin *Plugin, ui JobUI) {

	defer ui.JobDone(dir)

	var (
		err     error
		path    string
		out     []byte
		entries []string

		url      = plugin.url.String()
		stripDir = plugin.opts.stripDir
		sha1     = plugin.opts.sha1
	)

	if dryRun {
		entries, err = dryAcquire(dir, url, plugin.archive, stripDir, sha1)
	} else {
		err = acquire(dir, plugin.ext, url, plugin.archive, stripDir, sha1)
	}
	if err != nil {
		log.Printf("%s: %v", dir, err)
		return
	}
	if dryRun {
		ui.Print(dir, strings.Join(entries, "\n"))
	}

	//
	// no postUpdate hook? done.
	//
	if plugin.opts.postUpdate == "" {
		return
	}

	//
	// handle the .postupdate hook
	//
	path, err = expandPath(plugin.opts.postUpdate)
	if err != nil {
		log.Printf("%s: expanding .postupdate %q: %s", dir, plugin.opts.postUpdate, err)
		return
	}
	path = expandPathEnvironment(path, dir)

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

	if dryRun {
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
