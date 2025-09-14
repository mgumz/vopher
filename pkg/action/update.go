package action

import (
	"log"
	"os"
	"path"
	"path/filepath"

	"github.com/mgumz/vopher/pkg/action/update"
	"github.com/mgumz/vopher/pkg/archive"
	"github.com/mgumz/vopher/pkg/http"
	"github.com/mgumz/vopher/pkg/plugin"
	"github.com/mgumz/vopher/pkg/ui"
	"github.com/mgumz/vopher/pkg/utils"
)

// ActUpdateOpts holds options for action "update"
type ActUpdateOpts struct {
	Dir    string // dir to extract the plugins to
	Force  bool   // enforce acquire, even plugin exists
	DryRun bool   // don't extract, just show the files
	SHA1   string // checksum to check the downloaded file against
}

// Update fetches available updates for all given `plugins` and prints the
// result utilising the given ui.
func Update(plugins plugin.List, ui ui.UI, opts *ActUpdateOpts) {

	ui.Start()
	ui.Print("update", "started")

	var err error

	for _, plugin := range plugins {

		pluginFolder := filepath.Join(opts.Dir, plugin.Name)

		if _, err = os.Stat(pluginFolder); err == nil { // plugin_folder exists
			if !opts.Force {
				continue
			}
		}

		archiveName := filepath.Base(plugin.URL.Path)

		// apply heuristics - aka ""guess""
		if isArchive, _ := archive.IsSupportedArchive(archiveName); !isArchive {
			switch plugin.URL.Host {
			case "github.com":
				// NOTE: "master" as default one is a classic fallback, but
				// nowadays the primary branch is called "main" often. needs
				// to be tackled somewhen
				ref := utils.FirstNotEmpty(plugin.URL.Fragment, plugin.Opts.Branch, "master")
				remoteZip := ref + ".zip"
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
		go update.AcquireAndPostupdate(pluginFolder, opts.DryRun, plugin, ui)
	}

	ui.Wait()
	ui.Stop()
}
