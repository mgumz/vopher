package main

import (
	"fmt"
	"io"
	"path/filepath"
	"strings"
)

var supportedArchives = []string{
	".zip",
	".vba", ".vba.gz",
	".tar",
	".tar.gz",
	".tar.bz2", ".tar.bzip2",
}

// PluginArchive is the interface to show and extract entries
// from an archived version of a given plugin. this means zipped
// or tar.gzipped or vimballs
type PluginArchive interface {
	Extract(folder string, r io.Reader, skipDirs int) error
	Entries(r io.Reader, skipDirs int) ([]string, error)
}

var archiveGuesser = []func(string) PluginArchive{
	func(n string) PluginArchive {
		if strings.HasSuffix(n, ".zip") {
			return &ZipArchive{}
		} else if strings.HasSuffix(n, ".vba") {
			return &VimballArchive{}
		} else if strings.HasSuffix(n, ".vba.gz") {
			return &GzArchive{&VimballArchive{}}
		} else if strings.HasSuffix(n, ".tar") {
			return &TarArchive{}
		} else if strings.HasSuffix(n, ".tar.gz") {
			return &GzArchive{&TarArchive{}}
		} else if strings.HasSuffix(n, ".tar.bz2") || strings.HasSuffix(n, ".tar.bzip2") {
			return &BzipArchive{&TarArchive{}}
		}
		return nil
	},
}

// returns true/false if "name" is a supported archive type
// and the length of the suffix. eg, ".zip" yields 4, ".vba.gz"
// yields 7.
func isSupportedArchive(name string) (bool, int) {
	name = strings.ToLower(name)
	for i := range supportedArchives {
		if strings.HasSuffix(name, supportedArchives[i]) {
			return true, len(supportedArchives[i])
		}
	}
	return false, 0
}

func guessPluginArchive(name string) (PluginArchive, error) {
	n := strings.ToLower(name)
	for _, guessArchive := range archiveGuesser {
		archive := guessArchive(n)
		if archive != nil {
			return archive, nil
		}
	}
	return nil, fmt.Errorf("unsupported archive type for %q", name)
}

// strip away the leading 'strip_dirs' directories from 'name'. returns
// the stripped named AND a bool indicating, if the entry should be skipped
// because it's the root-direktory
//
//      name/      <- root-directory, will be stripped
//      name/a.vim
func stripArchiveEntry(name string, stripDirs int) (strippedName string, isRoot bool) {
	name = filepath.ToSlash(name)
	idx := indexByteN(name, '/', stripDirs)
	name = name[idx+1:]
	return name, (name == "")
}
