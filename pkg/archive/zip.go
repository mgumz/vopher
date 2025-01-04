package archive

import (
	"archive/zip"
	"bytes"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/mgumz/vopher/pkg/vopher"
)

const (
	// The bytes decompressed from a file read from the internets.
	// As of now, it is set to 100mb - which I fancy sufficient. I imagine
	// the usual plugin/-folder is rather in the low single digit megabyte
	// range. So, to give some headroom, I decided to increase by
	// two orders of magnitude.
	//
	// CWE-409: Potential DoS vulnerability via decompression bomb
	maxZipDecompressBytes = 1024 * 1024 * 100
)

// ZipArchive handles zip archive
type ZipArchive struct {
	GitCommit bool // if true: assume the .zip comment contains the git-commit
}

func init() {
	supportedArchives = append(supportedArchives, ".zip")
	archiveGuesser = append(archiveGuesser, func(n string) vopher.Archive {
		if strings.HasSuffix(n, ".zip") {
			return &ZipArchive{}
		}
		return nil
	})
}

func (za *ZipArchive) Extract(folder string, r io.Reader, stripDirs int) error {

	zfile, err := za.openReader(r)
	if err != nil {
		return err
	}

	for _, f := range zfile.File {

		oname, isRoot := stripArchiveEntry(f.Name, stripDirs)
		if isRoot {
			continue
		}

		oname = filepath.Join(folder, filepath.Clean(oname))

		if f.FileInfo().IsDir() {
			_ = os.MkdirAll(oname, 0700)
			continue
		}

		// TODO: call only if needed
		_ = os.MkdirAll(filepath.Dir(oname), 0700)

		zreader, err := f.Open()
		if err != nil {
			log.Println(oname, err)
		}
		ofile, err := os.Create(oname)
		if err != nil {
			log.Println(oname, err)
		}

		maxBytes := za.min(f.UncompressedSize64, maxZipDecompressBytes)

		_, err = io.CopyN(ofile, zreader, int64(maxBytes))
		if err != nil {
			log.Println(oname, err)
		}

		_ = ofile.Close()
		_ = zreader.Close()
	}

	// github stores the git-commit in the comment of the `.zip` file
	// so, we store a file called "github-commit" in the plugin-folder
	// to be able to check for updates
	if za.GitCommit && len(zfile.Comment) == 40 {
		name := filepath.Join(folder, "github-commit")
		file, err := os.Create(name)
		if err != nil {
			return err
		}
		defer (func() { _ = file.Close() })()
		_, _ = io.WriteString(file, zfile.Comment)
	}

	return nil
}

func (za *ZipArchive) Entries(r io.Reader, stripDirs int) ([]string, error) {

	var (
		entries    = make([]string, 0)
		zfile, err = za.openReader(r)
	)
	if err != nil {
		return nil, err
	}

	for _, f := range zfile.File {
		oname, isRoot := stripArchiveEntry(f.Name, stripDirs)
		if isRoot {
			continue
		}
		entries = append(entries, filepath.Clean(oname))
	}

	return entries, nil
}

func (*ZipArchive) openReader(r io.Reader) (*zip.Reader, error) {
	switch rt := r.(type) {
	default:
		buffer := bytes.NewBuffer(nil)
		if _, err := io.Copy(buffer, r); err != nil {
			return nil, err
		}
		br := bytes.NewReader(buffer.Bytes())
		return zip.NewReader(br, int64(buffer.Len()))
	case *os.File:
		fi, err := rt.Stat()
		if err != nil {
			return nil, err
		}
		return zip.NewReader(rt, fi.Size())
	}
}

func (*ZipArchive) min(a, b uint64) uint64 {
	if a < b {
		return a
	}
	return b
}
