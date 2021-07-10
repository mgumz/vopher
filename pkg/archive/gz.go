package archive

import (
	"compress/gzip"
	"io"

	"github.com/mgumz/vopher/pkg/vopher"
)

// GzArchive handles gzip archives
type GzArchive struct{ orig vopher.Archive }

// Extract ungzips the given archive `ga` into `folder`
func (ga *GzArchive) Extract(folder string, r io.Reader, skipDir int) error {
	gzreader, err := gzip.NewReader(r)
	if err != nil {
		return err
	}
	return ga.orig.Extract(folder, gzreader, skipDir)
}

// Entries lists the entries inside of given archive `ga`
func (ga *GzArchive) Entries(r io.Reader, skipDir int) ([]string, error) {
	gzreader, err := gzip.NewReader(r)
	if err != nil {
		return nil, err
	}
	return ga.orig.Entries(gzreader, skipDir)
}
