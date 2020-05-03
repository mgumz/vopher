package action

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"

	"github.com/mgumz/vopher/pkg/acquire"
	"github.com/mgumz/vopher/pkg/archive"
	"github.com/mgumz/vopher/pkg/http"
	"github.com/mgumz/vopher/pkg/plugin"
	"github.com/mgumz/vopher/pkg/ui"
	"github.com/mgumz/vopher/pkg/utils"
)

type ActUpdateOpts struct {
	Dir    string // dir to extract the plugins to
	Force  bool   // enforce acquire, even plugin exists
	DryRun bool   // don't extract, just show the files
	SHA1   string // checksum to check the downloaded file against
}

func Update(plugins plugin.List, ui ui.UI, opts *ActUpdateOpts) {

	ui.Start()

	var err error

	for _, plugin := range plugins {

		pluginFolder := filepath.Join(opts.Dir, plugin.Name)

		if _, err = os.Stat(pluginFolder); err == nil { // plugin_folder exists
			if !opts.Force {
				continue
			}
		}

		archiveName := filepath.Base(plugin.URL.Path)

		// apply heuristics aka ""guess""
		if isArchive, _ := archive.IsSupportedArchive(archiveName); !isArchive {
			switch plugin.URL.Host {
			case "github.com":
				remoteZip := utils.FirstNotEmpty(plugin.URL.Fragment, "master") + ".zip"
				plugin.URL.Path = path.Join(plugin.URL.Path, "archive", remoteZip)
				archiveName = filepath.Base(remoteZip)
				plugin.Ext, plugin.Archive = ".zip", &archive.ZipArchive{GitCommit: true}
			default:
				plugin.Ext, err = http.DetectFType(plugin.URL.String())
				if err != nil {
					log.Printf("error: %q: %s", plugin.URL, err)
					continue
				}
				archiveName += plugin.Ext
			}
		}

		if ok, suffixLen := archive.IsSupportedArchive(archiveName); ok {
			plugin.Ext = archiveName[len(archiveName)-suffixLen:]
		}

		if plugin.Archive == nil {
			plugin.Archive, err = archive.GuessArchive(archiveName)
			if err != nil {
				log.Printf("error: %q: not supported archive format", plugin.URL)
				continue
			}
		}

		ui.AddJob(pluginFolder)
		go acquireAndPostupdate(pluginFolder, opts.DryRun, plugin, ui)
	}

	ui.Wait()
	ui.Stop()
}

func acquireAndPostupdate(dir string, dryRun bool, plugin *plugin.Plugin, ui ui.UI) {

	defer ui.JobDone(dir)

	var (
		err     error
		path    string
		out     []byte
		entries []string

		url      = plugin.URL.String()
		stripDir = plugin.Opts.StripDir
		sha1     = plugin.Opts.SHA1
	)

	if dryRun {
		entries, err = acquire.DryAcquire(dir, url, plugin.Archive, stripDir, sha1)
	} else {
		err = acquire.Acquire(dir, plugin.Ext, url, plugin.Archive, stripDir, sha1)
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
	if plugin.Opts.PostUpdate == "" {
		return
	}

	//
	// handle the .postupdate hook
	//
	path, err = utils.ExpandPath(plugin.Opts.PostUpdate)
	if err != nil {
		log.Printf("%s: expanding .postupdate %q: %s", dir, plugin.Opts.PostUpdate, err)
		return
	}
	path = utils.ExpandPathEnvironment(path, dir)

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
			"VOPHER_NAME="+plugin.Name,
			"VOPHER_ARCHIVE="+plugin.Name+plugin.Ext,
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
