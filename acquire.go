package main

import (
	"bytes"
	"fmt"
	"os"
)

// fetch 'url' and extract it into 'base'. skip 'skip_dirs'
// leading directories in filenames in zip while extracting
// the contents.
func acquire(base, ext, url string, archive PluginArchive, skipDirs int, checkSha1 string) error {

	err := os.MkdirAll(base, 0777)

	if err != nil {
		return fmt.Errorf("mkdir %q: %s", base, err)
	}

	name := base + ext
	file, err := os.Create(name)
	if err != nil {
		return err
	}
	defer file.Close()

	if err = httpGET(file, url, checkSha1); err != nil {
		return err
	}

	file.Sync()
	file.Seek(0, 0)

	return archive.Extract(base, file, skipDirs)
}

// download 'url' and try to parse the zip-file. print out
// the files inside the zip while applying 'skip_dirs'.
//
// TODO: dryAcquire is not nice with ui 'oneline' (or future 'curses' based
// ones)
//
func dryAcquire(base, url string, archive PluginArchive, skipDirs int, checkSha1 string) ([]string, error) {

	buffer := bytes.NewBuffer(nil)
	if err := httpGET(buffer, url, checkSha1); err != nil {
		return nil, err
	}
	defer buffer.Reset()

	br := bytes.NewReader(buffer.Bytes())
	entries, err := archive.Entries(br, skipDirs)

	if err != nil {
		return nil, fmt.Errorf("gettting contents: %v", err)
	}
	if len(entries) == 0 {
		return nil, fmt.Errorf("empty archive for %q", url)
	}

	return entries, nil
}
