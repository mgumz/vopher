package main

import (
	"bytes"
	"crypto/sha1"
	"encoding/hex"
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
func acquire(base, ext, url string, archive PluginArchive, skip_dirs int, checkSha1 string) error {

	var (
		name = base
		err  = os.MkdirAll(name, 0777)
	)

	if err != nil {
		return fmt.Errorf("mkdir %q: %s", base, err)
	}
	if filepath.Ext(name) == "" {
		name += ext
	}
	if err = httpget(name, url, checkSha1); err != nil {
		return err
	}
	file, err := os.Open(name)
	if err != nil {
		return err
	}
	defer file.Close()

	return archive.Extract(base, file, skip_dirs)
}

// download 'url' and try to parse the zip-file. print out
// the files inside the zip while applying 'skip_dirs'.
//
// TODO: dry_acquire is not nice with ui 'oneline' (or future 'curses' based
// ones)
//
func dry_acquire(base, url string, archive PluginArchive, skip_dirs int, checkSha1 string) ([]string, error) {

	var (
		err     error
		resp    *http.Response
		buffer  *bytes.Buffer
		br      *bytes.Reader
		entries []string
	)

	if resp, err = http.Get(url); err != nil {
		return nil, err
	} else if resp.StatusCode != 200 {
		log.Println(resp)
		return nil, fmt.Errorf("%d for %q", resp.StatusCode, url)
	}
	defer resp.Body.Close()

	hasher := sha1.New()
	tee := io.TeeReader(resp.Body, hasher)

	buffer = bytes.NewBuffer(nil)
	if _, err = io.Copy(buffer, tee); err != nil {
		return nil, err
	}
	defer buffer.Reset()

	sha1Sum := hex.EncodeToString(hasher.Sum(nil))

	if checkSha1 != "" && checkSha1 != sha1Sum {
		return nil, fmt.Errorf("sha1 does not match: got %s, expected %s", sha1Sum, checkSha1)
	}
	br = bytes.NewReader(buffer.Bytes())
	entries, err = archive.Entries(br, skip_dirs)

	if err != nil {
		return nil, fmt.Errorf("gettting contents: %v", err)
	}
	if len(entries) == 0 {
		return nil, fmt.Errorf("empty archive for %q", url)
	}

	return entries, nil
}
