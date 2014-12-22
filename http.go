package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path"
	"strings"
)

// HEAD url, find out "Content-Disposition: attachment; filename=foo.EXT"
func httpdetect_ftype(url string) (string, error) {

	resp, err := http.Head(url)
	if err != nil {
		return "", err
	}
	resp.Body.Close()

	content := resp.Header.Get("Content-Disposition")
	if content == "" {
		return "", fmt.Errorf("can't detect filetype of %q", url)
	}

	const fn = "filename="
	idx := strings.Index(content, fn)
	if idx == -1 || len(content) == idx+len(fn) {
		return "", fmt.Errorf("invalid 'Content-Disposition' header for %q", url)
	}

	return path.Ext(content[idx+len(fn):]), nil
}

func httpget(out, url string) (err error) {

	var file *os.File
	var resp *http.Response

	if file, err = os.Create(out); err != nil {
		return err
	}
	defer file.Close()

	if resp, err = http.Get(url); err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		log.Println(resp)
		return fmt.Errorf("%d for %q", resp.StatusCode, url)
	}

	reader := io.Reader(resp.Body)
	/*
		if resp.ContentLength > 0 {
			progress := NewProgressTicker(resp.ContentLength)
			defer progress.Stop()
			go progress.Start(out, 2*time.Millisecond)
			reader = io.TeeReader(reader, progress)
		}
	*/

	_, err = io.Copy(file, reader)
	return err
}
