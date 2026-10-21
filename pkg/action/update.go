package action

import (
	"fmt"
	"log"
	"net/url"
	"os"
	"path/filepath"

	"github.com/mgumz/vopher/pkg/acquire"
	"github.com/mgumz/vopher/pkg/action/update"
	"github.com/mgumz/vopher/pkg/archive"
	"github.com/mgumz/vopher/pkg/http"
	"github.com/mgumz/vopher/pkg/plugin"
	"github.com/mgumz/vopher/pkg/ui"
	"github.com/mgumz/vopher/pkg/utils"
)

// ActUpdateOpts holds options for action "update"
type ActUpdateOpts struct {
	Dir     string // dir to extract the plugins to
	Force   bool   // enforce acquire, even plugin exists
	DryRun  bool   // don't extract, just show the files
	SHA1    string // checksum to check the downloaded file against
	Cache   string // dir to keep the downloaded archives in; "" = next to the plugins in Dir
	Offline bool   // never fetch, use the kept archives only
}

// Update fetches available updates for all given `plugins` and prints the
// result utilising the given ui.
func Update(plugins plugin.List, ui ui.UI, opts *ActUpdateOpts) {

	ui.Start()
	ui.Print("update", "started")

	// all jobs are added before the first one starts: resolving a source
	// may probe the hoster, so the total of the progress is known up front
	// and the probes run concurrently, limited per host by -http.jobs
	todo := make(map[string]*plugin.Plugin, len(plugins))
	for _, plugin := range plugins {
		pluginFolder := filepath.Join(opts.Dir, plugin.Name)
		if _, err := os.Stat(pluginFolder); err == nil { // plugin_folder exists
			if !opts.Force {
				continue
			}
		}
		todo[pluginFolder] = plugin
		ui.AddJob(pluginFolder)
	}

	for pluginFolder, plugin := range todo {
		go func() {
			src, err := resolveSource(plugin, pluginFolder, opts)
			if err != nil {
				log.Printf("error: %q: %s", plugin.URL, err)
				ui.JobDone(pluginFolder)
				return
			}
			update.AcquireAndPostupdate(pluginFolder, src, opts.DryRun, plugin, ui)
		}()
	}

	ui.Wait()
	ui.Stop()
}

// resolveSource figures out where the archive of plugin comes from and of
// which type it is; plugin.URL, .Ext and .Archive are set accordingly
func resolveSource(plugin *plugin.Plugin, pluginFolder string, opts *ActUpdateOpts) (acquire.Source, error) {

	archiveName := filepath.Base(plugin.URL.Path)
	stem := pluginFolder
	if opts.Cache != "" {
		stem = acquire.ArchiveStem(opts.Cache, plugin.URL)
	}
	immutable := false

	hoster, err := acquire.HosterFor(plugin)
	if err != nil {
		return acquire.Source{}, err
	}

	// apply heuristics - aka ""guess""
	if isArchive, _ := archive.IsSupportedArchive(archiveName); !isArchive {
		switch {
		case hoster != nil:
			ref := utils.FirstNotEmpty(plugin.URL.Fragment, plugin.Opts.Branch)
			immutable = acquire.IsCommit(ref)
			if ref == "" {
				ref, err = defaultBranch(hoster, plugin.URL, pluginFolder, opts)
				if err != nil {
					return acquire.Source{}, err
				}
			}
			remoteZip := hoster.ArchiveURL(plugin.URL, ref)
			plugin.URL = &remoteZip
			if opts.Cache != "" {
				stem = acquire.ArchiveStem(opts.Cache, plugin.URL)
			}
			archiveName = filepath.Base(plugin.URL.Path)
			plugin.Ext, plugin.Archive = ".zip", &archive.ZipArchive{GitCommit: true}
		default:
			plugin.Ext, err = detectExt(plugin.URL.String(), stem, opts.Offline)
			if err != nil {
				return acquire.Source{}, err
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
			return acquire.Source{}, fmt.Errorf("not supported archive format")
		}
	}

	return acquire.Source{
		URL:       plugin.URL.String(),
		Path:      acquire.WithExt(stem, plugin.Ext),
		SHA1:      plugin.Opts.SHA1,
		Immutable: immutable,
		Offline:   opts.Offline,
	}, nil
}

// detectExt asks the server for the archive type of url; offline, the type
// of the kept archive with the given stem is used
func detectExt(url, stem string, offline bool) (string, error) {
	if !offline {
		return http.DetectFType(url)
	}
	name, ok := acquire.FindArchive(stem)
	if !ok {
		return "", fmt.Errorf("not available offline")
	}
	_, suffixLen := archive.IsSupportedArchive(name)
	return name[len(name)-suffixLen:], nil
}

// defaultBranch asks the hoster for the first of acquire.DefaultBranches
// with an archive; offline, it takes the first of them with a kept archive
func defaultBranch(hoster acquire.Hoster, repo *url.URL, pluginFolder string, opts *ActUpdateOpts) (string, error) {
	if !opts.Offline {
		return acquire.ProbeBranch(hoster, repo, acquire.DefaultBranches...)
	}
	for _, branch := range acquire.DefaultBranches {
		stem := pluginFolder
		if opts.Cache != "" {
			archiveURL := hoster.ArchiveURL(repo, branch)
			stem = acquire.ArchiveStem(opts.Cache, &archiveURL)
		}
		if _, ok := acquire.FindArchive(stem); ok {
			return branch, nil
		}
	}
	return "", fmt.Errorf("not available offline")
}
