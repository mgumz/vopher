// +build lzma

package main

import (
	"io"
	"strings"

	"code.google.com/p/lzma"
)

func init() {
	supported_archives = append(supported_archives, ".tar.lzma")
	supported_archives = append(supported_archives, ".tar.xz")

	archive_guesser = append(archive_guesser, func(n string) PluginArchive {
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
