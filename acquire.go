package main

import (
	"archive/zip"
	"bytes"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
)

// fetch 'url' and extract it into 'base'. skip 'skip_dirs'
// leading directories in filenames in zip while extracting
// the contents.
func acquire(base, url string, skip_dirs int) error {

	if err := os.MkdirAll(base, 0777); err != nil {
		return fmt.Errorf("mkdir %q: %s", base, err)
	}

	name := base + ".zip"
	if err := httpget(name, url); err != nil {
		return err
	}
	zfile, err := zip.OpenReader(name)
	if err != nil {
		return err
	}
	defer zfile.Close()
	for _, f := range zfile.File {
		idx := index_byte_n(f.Name, '/', skip_dirs)

		oname := f.Name[idx+1:]

		// root-directory
		//   pname/      <- root-directory
		//   pname/a.vim
		if oname == "" {
			continue
		}

		oname = filepath.Join(base, filepath.Clean(oname))

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

// download 'url' and try to parse the zip-file. print out
// the files inside the zip while applying 'skip_dirs'.
//
// TODO: dry_acquire is not nice with ui 'oneline' (or future 'curses' based
// ones)
//
func dry_acquire(base, url string, skip_dirs int) error {

	var (
		err    error
		resp   *http.Response
		buffer *bytes.Buffer
		zfile  *zip.Reader
	)

	if resp, err = http.Get(url); err != nil {
		return err
	} else if resp.StatusCode != 200 {
		log.Println(resp)
		return fmt.Errorf("%d for %q", resp.StatusCode, url)
	}
	defer resp.Body.Close()

	buffer = bytes.NewBuffer(nil)
	if _, err = io.Copy(buffer, resp.Body); err != nil {
		return err
	}
	defer buffer.Reset()

	if zfile, err = zip.NewReader(bytes.NewReader(buffer.Bytes()), int64(buffer.Len())); err != nil {
		return err
	}

	for _, f := range zfile.File {

		idx := index_byte_n(f.Name, '/', skip_dirs)
		oname := f.Name[idx+1:]

		// root-directory
		//   pname/      <- root-directory
		//   pname/a.vim
		if oname == "" {
			continue
		}

		oname = filepath.Join(base, filepath.Clean(oname))

		fmt.Printf("%s\t%d\n", oname, f.FileHeader.UncompressedSize)
	}

	return nil
}
