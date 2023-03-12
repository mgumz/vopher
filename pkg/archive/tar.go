package archive

import (
	"archive/tar"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/mgumz/vopher/pkg/utils"
	"github.com/mgumz/vopher/pkg/vopher"
)

const (
	// the bytes decompressed from a file read from the internets.
	// as of now it is set to 50mb per entry in a tar-archive. i imagine
	// the usual plugin/-folder is rather in the low single digit megabyte
	// range. so, to give some headroom, i decided to increase by
	// one order of magnitude.
	//
	// CWE-409: Potential DoS vulnerability via decompression bomb
	maxTarDecompressEntryBytes = 1024 * 1024 * 50
)

// TarArchive handles tar archives
type TarArchive struct{}

func init() {

	suffixes := []string{
		".tar",
		".tgz", ".tar.gz",
		".tar.bz2", ".tar.bzip2",
	}
	supportedArchives = append(supportedArchives, suffixes...)

	archiveGuesser = append(archiveGuesser, func(n string) vopher.Archive {

		if utils.StringHasSuffix(n, []string{".tar"}) {
			return &TarArchive{}
		} else if utils.StringHasSuffix(n, []string{".tar.gz", ".tgz"}) {
			return &GzArchive{&TarArchive{}}
		} else if utils.StringHasSuffix(n, []string{".tar.bz2", ".tar.bzip2"}) {
			return &BzipArchive{&TarArchive{}}
		}
		return nil
	})

}

// Extract untars the given archive `ta` into `folder`
func (ta *TarArchive) Extract(folder string, r io.Reader, stripDirs int) error {
	_, err := ta.handle(folder, r, stripDirs, tarExtractEntry)
	return err
}

// Entries lists the entries inside of given archive `ta`
func (ta *TarArchive) Entries(r io.Reader, stripDirs int) ([]string, error) {
	return ta.handle("", r, stripDirs, tarIgnoreEntry)
}

// small helper to operate on a tar-entry. reader r points directly
// to the data for 'name' in the tar file.
type tarExtractFunc func(name string, r io.Reader, maxBytes int64) error

// handle all file-like entries in the tar represented by 'r' due the 'extract'
// function.
// TODO: make sure "file-like" is the correct criteria.
func (ta *TarArchive) handle(folder string, r io.Reader, stripDirs int, extract tarExtractFunc) ([]string, error) {

	var (
		entries = make([]string, 0)
		reader  = tar.NewReader(r)
	)

	for {
		header, err := reader.Next()
		if err == io.EOF {
			break
		} else if err != nil {
			return nil, err
		}

		if header.Typeflag != tar.TypeReg {
			continue
		}

		fi := header.FileInfo()
		if fi.IsDir() {
			// TODO: decide, if we skip dirs or not
			continue
		} else if strings.HasPrefix(header.Name, "/") {
			return nil, fmt.Errorf("entry with absolute filename %q", header.Name)
		}

		oname, isRoot := stripArchiveEntry(header.Name, stripDirs)
		if isRoot {
			continue
		}
		entries = append(entries, oname)

		maxBytes := ta.min(header.Size, maxTarDecompressEntryBytes)

		err = extract(filepath.Join(folder, oname), reader, maxBytes)
		if err != nil {
			return nil, err
		}
	}
	return entries, nil
}

func tarExtractEntry(name string, r io.Reader, maxBytes int64) error {
	if err := os.MkdirAll(filepath.Dir(name), 0700); err != nil {
		return err
	}
	file, err := os.Create(name)
	if err != nil {
		return err
	}
	_, err = io.CopyN(file, r, maxBytes)
	return err
}
func tarIgnoreEntry(name string, r io.Reader, maxBytes int64) error {
	_, err := io.CopyN(io.Discard, r, maxBytes)
	return err
}

func (*TarArchive) min(a, b int64) int64 {
	if a < b {
		return a
	}
	return b
}
