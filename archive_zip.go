package main

import (
	"archive/zip"
	"bytes"
	"io"
	"log"
	"os"
	"path/filepath"
)

type ZipArchive struct{}

func (za *ZipArchive) Extract(folder string, r io.Reader, strip_dirs int) error {

	zfile, err := zip_openReader(r)
	if err != nil {
		return err
	}

	for _, f := range zfile.File {

		oname, is_root := StripArchiveEntry(f.Name, strip_dirs)
		if is_root {
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

	return nil
}

func (*ZipArchive) Entries(r io.Reader, strip_dirs int) ([]string, error) {

	var (
		entries    = make([]string, 0)
		zfile, err = zip_openReader(r)
	)
	if err != nil {
		return nil, err
	}

	for _, f := range zfile.File {
		oname, is_root := StripArchiveEntry(f.Name, strip_dirs)
		if is_root {
			continue
		}
		entries = append(entries, filepath.Clean(oname))
	}

	return entries, nil
}

func zip_openReader(r io.Reader) (*zip.Reader, error) {
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
