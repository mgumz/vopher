// +build lzma

package main

import (
	"io"
	"strings"

	"code.google.com/p/lzma"
)

func init() {
	supportedArchives = append(supportedArchives, ".tar.lzma", ".tar.xz")

	archiveGuesser = append(archiveGuesser, func(n string) PluginArchive {
		if strings.HasSuffix(n, ".tar.lzma") {
			return &LzmaArchive{&TarArchive{}}
		} else if strings.HasSuffix(n, ".tar.xz") {
			return &LzmaArchive{&TarArchive{}}
		}
		return nil
	})
}

// wrapper to decompress lzma
type lzmaArchive struct{ orig PluginArchive }

func (la *lzmaArchive) extract(folder string, r io.Reader, skipDir int) error {
	lzmaReader := lzma.NewReader(r)
	defer lzmaReader.Close()
	return la.orig.extract(folder, lzmaReader, skipDir)
}

func (la *lzmaArchive) entries(r io.Reader, skipDir int) ([]string, error) {
	lzmaReader := lzma.NewReader(r)
	defer lzmaReader.Close()
	return la.orig.entries(lzmaReader, skipDir)
}
