package main

import (
	"archive/zip"
	"bytes"
	"io"
	"log"
	"os"
	"path/filepath"
)

// ZipArchive handles zip archive
type ZipArchive struct {
	gitCommit bool // if true: assume the .zip comment contains the git-commit
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
			os.MkdirAll(oname, 0777)
			continue
		}

		// TODO: call only if needed
		os.MkdirAll(filepath.Dir(oname), 0777)

		zreader, err := f.Open()
		if err != nil {
			log.Println(oname, err)
		}
		ofile, err := os.Create(oname)
		if err != nil {
			log.Println(oname, err)
		}
		_, err = io.Copy(ofile, zreader)
		if err != nil {
			log.Println(oname, err)
		}

		ofile.Close()
		zreader.Close()
	}

	// github stores the git-commit in the comment of the .zip file
	// so, we store a file called "github-commit" in the plugin-folder
	// to be able to check for updates
	if za.gitCommit && len(zfile.Comment) == 40 {
		name := filepath.Join(folder, "github-commit")
		file, err := os.Create(name)
		if err != nil {
			return err
		}
		defer file.Close()
		io.WriteString(file, zfile.Comment)
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
