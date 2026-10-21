package http

import (
	"crypto/sha1" /* #nosec */
	"encoding/hex"
	"fmt"
	"io"
	"log"
	nethttp "net/http"
	"path"
	"strings"
)

// DetectFType issues HEAD to url, examines "Content-Disposition: attachment; filename=foo.EXT"
func DetectFType(url string) (string, error) {

	resp, done, err := Request(nethttp.MethodHead, url)
	if err != nil {
		return "", err
	}
	done()

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

// Get fetches URL by streaming the data to io.Writer w
// and comparing the downloaded data to checkSha1
func Get(w io.Writer, url, checkSha1 string) error {
	_, _, err := GetIfChanged(w, url, checkSha1, Validators{})
	return err
}

// Validators identify the version of a resource a client already has, see
// RFC 9110, section 8.8
type Validators struct {
	ETag         string `json:"etag,omitempty"`
	LastModified string `json:"last_modified,omitempty"`
}

// GetIfChanged is Get, but sends the given validators along. If the server
// answers "304 Not Modified", nothing is written to w and notModified is
// true. Returns the validators of the fetched resource.
func GetIfChanged(w io.Writer, url, checkSha1 string, have Validators) (got Validators, notModified bool, err error) {

	header := nethttp.Header{}
	if have.ETag != "" {
		header.Set("If-None-Match", have.ETag)
	}
	if have.LastModified != "" {
		header.Set("If-Modified-Since", have.LastModified)
	}

	resp, done, err := RequestWithHeader(nethttp.MethodGet, url, header)
	if err != nil {
		return got, false, err
	}
	defer done()

	got = Validators{
		ETag:         resp.Header.Get("ETag"),
		LastModified: resp.Header.Get("Last-Modified"),
	}

	if resp.StatusCode == nethttp.StatusNotModified && (have.ETag != "" || have.LastModified != "") {
		return got, true, nil
	}
	if resp.StatusCode != nethttp.StatusOK {
		log.Println(resp)
		return got, false, fmt.Errorf("%d for %q", resp.StatusCode, url)
	}

	reader := io.Reader(resp.Body)
	hasher := sha1.New() /* #nosec */

	if checkSha1 != "" {
		reader = io.TeeReader(reader, hasher)
	}

	/*
		// NOTE: idea to report sub-progress, but maybe not worth the
		// effort since plugins are really small

		if resp.ContentLength > 0 {
			progress := NewProgressTicker(resp.ContentLength)
			defer progress.Stop()
			go progress.Start(out, 2*time.Millisecond)
			reader = io.TeeReader(reader, progress)
		}
	*/

	if _, err = io.Copy(w, reader); err != nil {
		return got, false, err
	}

	sha1Sum := hex.EncodeToString(hasher.Sum(nil))
	if checkSha1 != "" && checkSha1 != sha1Sum {
		return got, false, fmt.Errorf("sha1 does not match: got %s, expected %s", sha1Sum, checkSha1)
	}

	return got, false, nil
}
