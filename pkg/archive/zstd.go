//go:build zstd

package archive

import (
	"io"
    "strings"

	"github.com/klauspost/compress/zstd"

	"github.com/mgumz/vopher/pkg/vopher"
)

func init() {
	supportedArchives = append(supportedArchives, ".tar.zst")

	archiveGuesser = append(archiveGuesser, func(n string) vopher.Archive {
		if strings.HasSuffix(n, ".tar.zst") {
			return &ZstdArchive{&TarArchive{}}
		}
		return nil
	})
}

type ZstdArchive struct{ orig vopher.Archive }

func (za *ZstdArchive) Extract(folder string, r io.Reader, skipDir int) error {
	zr, _ := zstd.NewReader(r)
	defer zr.Close()
	return za.orig.Extract(folder, zr, skipDir)
}

func (za *ZstdArchive) Entries(r io.Reader, skipDir int) ([]string, error) {
	zr, _ := zstd.NewReader(r)
	defer zr.Close()
	return za.orig.Entries(zr, skipDir)
}
