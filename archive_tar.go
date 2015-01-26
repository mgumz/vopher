package main

import (
	"archive/tar"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
)

type TarArchive struct{}

func (ta *TarArchive) Extract(folder string, r io.Reader, strip_dirs int) error {
	_, err := ta.handle(folder, r, strip_dirs, tar_extract_entry)
	return err
}

func (ta *TarArchive) Entries(r io.Reader, strip_dirs int) ([]string, error) {
	return ta.handle("", r, strip_dirs, tar_ignore_entry)
}

// small helper to operate on a tar-entry. reader r points directly
// to the data for 'name' in the tar file.
type tar_efunc func(name string, r io.Reader) error

// handle all file-like entries in the tar represented by 'r' due the 'extract'
// function.
// TODO: make sure "file-like" is the correct criteria.
func (ta *TarArchive) handle(folder string, r io.Reader, strip_dirs int, extract tar_efunc) ([]string, error) {

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

		if header.Typeflag != tar.TypeReg && header.Typeflag != tar.TypeRegA {
			continue
		}

		fi := header.FileInfo()
		if fi.IsDir() {
			// TODO: decide, if we skip dirs or not
			continue
		} else if header.Name != "" && header.Name[0] == '/' {
			return nil, fmt.Errorf("entry with absolute filename %q", header.Name)
		}

		oname, is_root := StripArchiveEntry(header.Name, strip_dirs)
		if is_root {
			continue
		}
		entries = append(entries, oname)
		if err := extract(filepath.Join(folder, oname), reader); err != nil {
			return nil, err
		}
	}
	return entries, nil
}

func tar_extract_entry(name string, r io.Reader) error {
	if err := os.MkdirAll(filepath.Dir(name), 0777); err != nil {
		return err
	}
	file, err := os.Create(name)
	if err != nil {
		return err
	}
	_, err = io.Copy(file, r)
	return err
}
func tar_ignore_entry(name string, r io.Reader) error {
	_, err := io.Copy(ioutil.Discard, r)
	return err
}
