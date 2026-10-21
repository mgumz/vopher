package acquire

import (
	"archive/zip"
	"bytes"
	"crypto/sha1" /* #nosec */
	"encoding/hex"
	nethttp "net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"sync/atomic"
	"testing"

	"github.com/mgumz/vopher/pkg/archive"
	vhttp "github.com/mgumz/vopher/pkg/http"
)

func testZip(t *testing.T, content string) []byte {
	t.Helper()
	buf := bytes.NewBuffer(nil)
	zw := zip.NewWriter(buf)
	w, err := zw.Create("plugin-master/plugin/a.vim")
	if err != nil {
		t.Fatal(err)
	}
	_, _ = w.Write([]byte(content))
	if err = zw.Close(); err != nil {
		t.Fatal(err)
	}
	return buf.Bytes()
}

// testServer serves data with an ETag and answers matching conditional
// requests with 304
func testServer(t *testing.T, data []byte) (*httptest.Server, *atomic.Int32, *atomic.Int32) {
	t.Helper()
	const etag = `"v1"`
	hits, notModified := &atomic.Int32{}, &atomic.Int32{}
	srv := httptest.NewServer(nethttp.HandlerFunc(func(w nethttp.ResponseWriter, r *nethttp.Request) {
		hits.Add(1)
		w.Header().Set("ETag", etag)
		if r.Header.Get("If-None-Match") == etag {
			notModified.Add(1)
			w.WriteHeader(nethttp.StatusNotModified)
			return
		}
		_, _ = w.Write(data)
	}))
	t.Cleanup(srv.Close)
	return srv, hits, notModified
}

func mustExtracted(t *testing.T, base, want string) {
	t.Helper()
	got, err := os.ReadFile(filepath.Join(base, "plugin", "a.vim"))
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}

func TestAcquireCache(t *testing.T) {
	data := testZip(t, "v1")
	srv, hits, notModified := testServer(t, data)
	dir := t.TempDir()
	src := Source{URL: srv.URL + "/plugin.zip", Path: filepath.Join(dir, "cache", "plugin.zip")}
	acquire := func(src Source) error {
		base := filepath.Join(dir, "plugin")
		_ = os.RemoveAll(base)
		if err := Acquire(base, src, &archive.ZipArchive{}, 1); err != nil {
			return err
		}
		mustExtracted(t, base, "v1")
		return nil
	}

	if err := acquire(src); err != nil {
		t.Fatal(err)
	}
	if hits.Load() != 1 {
		t.Fatalf("first acquire: %d requests", hits.Load())
	}
	if _, err := os.Stat(src.Path + MetaSuffix); err != nil {
		t.Fatalf("no meta file: %s", err)
	}

	// kept archive + etag: conditional request, 304, extracted from cache
	if err := acquire(src); err != nil {
		t.Fatal(err)
	}
	if hits.Load() != 2 || notModified.Load() != 1 {
		t.Fatalf("conditional acquire: %d requests, %d not modified", hits.Load(), notModified.Load())
	}

	// no request at all
	sum := sha1.Sum(data) /* #nosec */
	for name, s := range map[string]Source{
		"immutable":    {URL: src.URL, Path: src.Path, Immutable: true},
		"offline":      {URL: src.URL, Path: src.Path, Offline: true},
		"sha1 matches": {URL: src.URL, Path: src.Path, SHA1: hex.EncodeToString(sum[:])},
	} {
		before := hits.Load()
		if err := acquire(s); err != nil {
			t.Fatalf("%s: %s", name, err)
		}
		if hits.Load() != before {
			t.Errorf("%s: %d requests, want none", name, hits.Load()-before)
		}
	}
}

func TestAcquireSHA1MismatchIgnoresETag(t *testing.T) {
	data := testZip(t, "v1")
	srv, _, notModified := testServer(t, data)
	dir := t.TempDir()
	src := Source{URL: srv.URL + "/plugin.zip", Path: filepath.Join(dir, "plugin.zip")}

	if err := os.WriteFile(src.Path, testZip(t, "stale"), 0600); err != nil {
		t.Fatal(err)
	}
	writeMeta(src.Path, vhttp.Validators{ETag: `"v1"`})
	sum := sha1.Sum(data) /* #nosec */
	src.SHA1 = hex.EncodeToString(sum[:])

	base := filepath.Join(dir, "plugin")
	if err := Acquire(base, src, &archive.ZipArchive{}, 1); err != nil {
		t.Fatal(err)
	}
	if notModified.Load() != 0 {
		t.Fatal("stale archive was confirmed by a 304")
	}
	mustExtracted(t, base, "v1")
}

func TestAcquireOfflineMissing(t *testing.T) {
	dir := t.TempDir()
	src := Source{URL: "http://127.0.0.1:1/plugin.zip", Path: filepath.Join(dir, "plugin.zip"), Offline: true}
	if err := Acquire(filepath.Join(dir, "plugin"), src, &archive.ZipArchive{}, 1); err == nil {
		t.Fatal("expected an error for a missing archive")
	}
}

func TestIsCommit(t *testing.T) {
	gh := Github{}
	for ref, want := range map[string]bool{
		"913fff1c": true,
		"913fff1cea3aa1a08a360a494fa05555e59147f5": true,
		"913fff":    false,
		"v2.1":      false,
		"master":    false,
		"deadbeefx": false,
		"":          false,
	} {
		if got := gh.IsCommit(ref); got != want {
			t.Errorf("%q: got %v, want %v", ref, got, want)
		}
	}
}
