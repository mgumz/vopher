// +build lzma

package main

import (
	"io"
	"strings"

	"github.com/smira/lzma"
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
type LzmaArchive struct{ orig PluginArchive }

func (la *LzmaArchive) Extract(folder string, r io.Reader, skipDir int) error {
	lzmaReader := lzma.NewReader(r)
	defer lzmaReader.Close()
	return la.orig.Extract(folder, lzmaReader, skipDir)
}

func (la *LzmaArchive) Entries(r io.Reader, skipDir int) ([]string, error) {
	lzmaReader := lzma.NewReader(r)
	defer lzmaReader.Close()
	return la.orig.Entries(lzmaReader, skipDir)
}
