package main

import (
	"compress/bzip2"
	"io"
)

// BzipArchive handles bzip2 archives
type BzipArchive struct{ orig PluginArchive }

func (ba *BzipArchive) Extract(folder string, r io.Reader, skipDir int) error {
	bzreader := bzip2.NewReader(r)
	return ba.orig.Extract(folder, bzreader, skipDir)
}

func (ba *BzipArchive) Entries(r io.Reader, skipDir int) ([]string, error) {
	bzreader := bzip2.NewReader(r)
	return ba.orig.Entries(bzreader, skipDir)
}
