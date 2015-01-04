package main

import (
	"archive/zip"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
)

// fetch 'url' and extract it into 'base'. skip 'skip_dirs'
// leading directories in filenames in zip while extracting
// the contents
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
