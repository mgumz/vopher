package vopher

import (
	"io"
)

// Archive is the interface to show and extract entries
// from an archived version of a given plugin. this means zipped
// or tar.gzipped or vimballs
type Archive interface {
	Extract(folder string, r io.Reader, skipDirs int) error
	Entries(r io.Reader, skipDirs int) ([]string, error)
}
