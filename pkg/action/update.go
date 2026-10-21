package action

import (
	"log"
	"os"
	"path/filepath"

	"github.com/mgumz/vopher/pkg/acquire"
	"github.com/mgumz/vopher/pkg/action/update"
	"github.com/mgumz/vopher/pkg/archive"
	"github.com/mgumz/vopher/pkg/http"
	"github.com/mgumz/vopher/pkg/plugin"
	"github.com/mgumz/vopher/pkg/ui"
)

// ActUpdateOpts holds options for action "update"
type ActUpdateOpts struct {
	Dir    string // dir to extract the plugins to
	Force  bool   // enforce acquire, even plugin exists
	DryRun bool   // don't extract, just show the files
	SHA1   string // checksum to check the downloaded file against
	Cache  string // dir to keep the downloaded archives in; "" = next to the plugins in Dir
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
				plugin.URL = acquire.Github{}.ArchiveURL(plugin.URL, plugin.Opts.Branch)
				archiveName = filepath.Base(plugin.URL.Path)
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

		archivePath := pluginFolder + plugin.Ext
		if opts.Cache != "" {
			archivePath = acquire.ArchivePath(opts.Cache, plugin.URL, plugin.Ext)
		}

		ui.AddJob(pluginFolder)
		go update.AcquireAndPostupdate(pluginFolder, archivePath, opts.DryRun, plugin, ui)
	}

	ui.Wait()
	ui.Stop()
}
