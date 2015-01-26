package main

import (
	"compress/bzip2"
	"compress/gzip"
	"fmt"
	"io"
	"path/filepath"
	"strings"

	"code.google.com/p/lzma"
)

var supported_archives = [...]string{
	".zip",
	".vba", ".vba.gz",
	".tar",
	".tar.gz",
	".tar.lzma",
	".tar.bz2", ".tar.bzip2",
	".tar.xz",
}

// returns true/false if "name" is a supported archive type
// and the length of the suffix. eg, ".zip" yields 4, ".vba.gz"
// yields 7.
func IsSupportedArchive(name string) (bool, int) {
	name = strings.ToLower(name)
	for i := range supported_archives {
		if strings.HasSuffix(name, supported_archives[i]) {
			return true, len(supported_archives[i])
		}
	}
	return false, 0
}

// a PluginArchive is the interface to show and extract entries
// from an archived version of a given plugin. this means zipped
// or tar.gzipped or vimballs
type PluginArchive interface {
	Extract(folder string, r io.Reader, skip_dirs int) error
	Entries(r io.Reader, skip_dirs int) ([]string, error)
}

func GuessPluginArchive(name string) (PluginArchive, error) {
	n := strings.ToLower(name)
	if strings.HasSuffix(n, ".zip") {
		return &ZipArchive{}, nil
	} else if strings.HasSuffix(n, ".vba") {
		return &VimballArchive{}, nil
	} else if strings.HasSuffix(n, ".vba.gz") {
		return &GzArchive{&VimballArchive{}}, nil
	} else if strings.HasSuffix(n, ".tar") {
		return &TarArchive{}, nil
	} else if strings.HasSuffix(n, ".tar.gz") {
		return &GzArchive{&TarArchive{}}, nil
	} else if strings.HasSuffix(n, ".tar.bz2") || strings.HasSuffix(n, ".tar.bzip2") {
		return &BzipArchive{&TarArchive{}}, nil
	} else if strings.HasSuffix(n, ".tar.xz") {
		return &LzmaArchive{&TarArchive{}}, nil
	}
	return nil, fmt.Errorf("unsupported archive type for %q\n", name)
}

// wrapper to decompress gzip
type GzArchive struct{ orig PluginArchive }

func (ga *GzArchive) Extract(folder string, r io.Reader, skip_dir int) error {
	gzreader, err := gzip.NewReader(r)
	if err != nil {
		return err
	}
	return ga.orig.Extract(folder, gzreader, skip_dir)
}

func (ga *GzArchive) Entries(r io.Reader, skip_dir int) ([]string, error) {
	gzreader, err := gzip.NewReader(r)
	if err != nil {
		return nil, err
	}
	return ga.orig.Entries(gzreader, skip_dir)
}

// wrapper to decompress bzip2
type BzipArchive struct{ orig PluginArchive }

func (ba *BzipArchive) Extract(folder string, r io.Reader, skip_dir int) error {
	bzreader := bzip2.NewReader(r)
	return ba.orig.Extract(folder, bzreader, skip_dir)
}

func (ba *BzipArchive) Entries(r io.Reader, skip_dir int) ([]string, error) {
	bzreader := bzip2.NewReader(r)
	return ba.orig.Entries(bzreader, skip_dir)
}

// wrapper to decompress lzma
type LzmaArchive struct{ orig PluginArchive }

func (la *LzmaArchive) Extract(folder string, r io.Reader, skip_dir int) error {
	lzma_reader := lzma.NewReader(r)
	defer lzma_reader.Close()
	return la.orig.Extract(folder, lzma_reader, skip_dir)
}

func (la *LzmaArchive) Entries(r io.Reader, skip_dir int) ([]string, error) {
	lzma_reader := lzma.NewReader(r)
	defer lzma_reader.Close()
	return la.Entries(lzma_reader, skip_dir)
}

// strip away the leading 'strip_dirs' directories from 'name'. returns
// the stripped named AND a bool indicating, if the entry should be skipped
// because it's the root-direktory
//
//      name/      <- root-directory, will be stripped
//      name/a.vim
func StripArchiveEntry(name string, strip_dirs int) (stripped_name string, is_root bool) {
	name = filepath.ToSlash(name)
	idx := index_byte_n(name, '/', strip_dirs)
	name = name[idx+1:]
	return name, (name == "")
}
