package action

import (
	"errors"
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/mgumz/vopher/pkg/acquire"
	"github.com/mgumz/vopher/pkg/plugin"
)

// Prune removes all given plugins from the disk. With `all`, archives in
// `cacheDir` not belonging to any of the plugins are removed as well.
func Prune(plugins plugin.List, base, cacheDir string, force, all bool) {

	dir, err := os.Open(base) // #nosec G304
	if err != nil {
		log.Println(err)
		return
	}
	defer dir.Close()

	entries, err := dir.Readdir(-1)
	if err != nil {
		log.Println(err)
		return
	}

	for i := range entries {

		name := filepath.Base(entries[i].Name())

		// spare plugin.zip from pruning
		if !all && strings.HasSuffix(name, ".zip") {
			if _, isPlugin := plugins[name[:len(name)-4]]; isPlugin {
				continue
			}
		}

		if _, isPlugin := plugins[name]; !isPlugin {
			pruneEntry(filepath.Join(base, name), force)
		}
	}

	if all && cacheDir != "" {
		pruneCache(plugins, cacheDir, force)
	}
}

func pruneEntry(name string, force bool) {
	suffix := "dry-run."
	fmt.Printf("prune %q: ", name)
	if force {
		suffix = "done."
		if err := os.RemoveAll(name); err != nil {
			suffix = err.Error()
		}
	}
	fmt.Println(suffix)
}

// pruneCache removes all files in cacheDir which are not the archive (or its
// meta file) of one of the plugins, see acquire.ArchiveStem
func pruneCache(plugins plugin.List, cacheDir string, force bool) {

	stems := make([]string, 0, len(plugins))
	for _, p := range plugins {
		u := p.URL
		if u.Host == "github.com" {
			u = acquire.Github{}.PluginArchiveURL(p)
		}
		stems = append(stems, acquire.ArchiveStem(cacheDir, u))
	}

	dirs := []string{}
	err := filepath.WalkDir(cacheDir, func(name string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			dirs = append(dirs, name)
			return nil
		}
		archiveName := strings.TrimSuffix(name, acquire.MetaSuffix)
		isArchive := func(stem string) bool { return acquire.IsArchiveOf(archiveName, stem) }
		if !slices.ContainsFunc(stems, isArchive) {
			pruneEntry(name, force)
		}
		return nil
	})
	if err != nil && !errors.Is(err, fs.ErrNotExist) {
		log.Println(err)
	}

	if !force {
		return
	}
	// deepest first; fails for non-empty dirs, which is intended
	for i := len(dirs) - 1; i > 0; i-- {
		_ = os.Remove(dirs[i])
	}
}
