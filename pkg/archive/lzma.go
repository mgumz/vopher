//go:build lzma

package archive

import (
	"io"
	"strings"

	"github.com/ulikunitz/xz/lzma"

	"github.com/mgumz/vopher/pkg/vopher"
)

func init() {
	supportedArchives = append(supportedArchives, ".tar.lzma", ".tar.xz")

	archiveGuesser = append(archiveGuesser, func(n string) vopher.Archive {
		if strings.HasSuffix(n, ".tar.lzma") {
			return &LzmaArchive{&TarArchive{}}
		} else if strings.HasSuffix(n, ".tar.xz") {
			return &LzmaArchive{&TarArchive{}}
		}
		return nil
	})
}

type LzmaArchive struct{ orig vopher.Archive }

func (la *LzmaArchive) Extract(folder string, r io.Reader, skipDir int) error {
	lzmaReader, err := lzma.NewReader(r)
	if err != nil {
		return err
	}
	return la.orig.Extract(folder, lzmaReader, skipDir)
}

func (la *LzmaArchive) Entries(r io.Reader, skipDir int) ([]string, error) {
	lzmaReader, err := lzma.NewReader(r)
	if err != nil {
		return nil, err
	}
	return la.orig.Entries(lzmaReader, skipDir)
}
