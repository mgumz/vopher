package main

import (
	"compress/gzip"
	"io"
)

// GzArchive handles gzip archives
type GzArchive struct{ orig PluginArchive }

func (ga *GzArchive) Extract(folder string, r io.Reader, skipDir int) error {
	gzreader, err := gzip.NewReader(r)
	if err != nil {
		return err
	}
	return ga.orig.Extract(folder, gzreader, skipDir)
}

func (ga *GzArchive) Entries(r io.Reader, skipDir int) ([]string, error) {
	gzreader, err := gzip.NewReader(r)
	if err != nil {
		return nil, err
	}
	return ga.orig.Entries(gzreader, skipDir)
}
