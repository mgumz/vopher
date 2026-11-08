// Copyright 2026 Mathias Gumz. All rights reserved. Use of this source code
// is governed by a BSD-style license that can be found in the LICENSE file.

//go:build ignore

// pack writes a release archive to stdout, the same on every build host:
//
//	go run ./tools/pack.go a b c | gzip -9 > all.tar.gz
//	go run ./tools/pack.go -zip a b c > all.zip
//
// a ustar tar by default, a zip with -zip. every file goes in under its
// base name, owned by 0:0 without user or group names, 0755 if it is
// executable, 0644 otherwise.
//
// why not the host's tar and zip: they put the account of the build host
// and the modes of the working copy (0600, 0700) into the archive, and no
// set of flags fixes both on every host. GNU tar takes --owner, --group
// and --mode; the bsdtar of macOS has no --mode and names the owner with
// --uid, --gid, --uname and --gname; zip has no flag for either. a
// release is built on whatever host is at hand, so the archive is
// written here instead.
package main

import (
	"archive/tar"
	"archive/zip"
	"compress/flate"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"
)

func main() {

	asZip := flag.Bool("zip", false, "write a zip instead of a tar")
	flag.Parse()

	write := writeTar
	if *asZip {
		write = writeZip
	}
	if err := write(os.Stdout, flag.Args()); err != nil {
		fmt.Fprintf(os.Stderr, "error: pack: %v\n", err)
		os.Exit(1)
	}
}

// entry opens 'name' and says how it goes into the archive
func entry(name string) (f *os.File, size int64, mode os.FileMode, mtime time.Time, err error) {

	if f, err = os.Open(name); err != nil {
		return
	}
	fi, err := f.Stat()
	switch {
	case err != nil:
	case !fi.Mode().IsRegular():
		err = fmt.Errorf("%q is not a regular file", name)
	default:
		mode = 0o644
		if fi.Mode()&0o111 != 0 {
			mode = 0o755
		}
		return f, fi.Size(), mode, fi.ModTime().UTC().Truncate(time.Second), nil
	}
	f.Close()
	return nil, 0, 0, time.Time{}, err
}

func writeTar(w io.Writer, files []string) error {

	tw := tar.NewWriter(w)
	for _, name := range files {
		f, size, mode, mtime, err := entry(name)
		if err != nil {
			return err
		}
		err = tw.WriteHeader(&tar.Header{
			Typeflag: tar.TypeReg,
			Name:     filepath.Base(name),
			Size:     size,
			Mode:     int64(mode),
			ModTime:  mtime,
			Format:   tar.FormatUSTAR,
		})
		if err == nil {
			_, err = io.Copy(tw, f)
		}
		f.Close()
		if err != nil {
			return fmt.Errorf("%q: %v", name, err)
		}
	}
	return tw.Close()
}

func writeZip(w io.Writer, files []string) error {

	zw := zip.NewWriter(w)
	zw.RegisterCompressor(zip.Deflate, func(w io.Writer) (io.WriteCloser, error) {
		return flate.NewWriter(w, flate.BestCompression)
	})
	for _, name := range files {
		f, _, mode, mtime, err := entry(name)
		if err != nil {
			return err
		}
		hdr := &zip.FileHeader{Name: filepath.Base(name), Method: zip.Deflate, Modified: mtime}
		hdr.SetMode(mode)
		zf, err := zw.CreateHeader(hdr)
		if err == nil {
			_, err = io.Copy(zf, f)
		}
		f.Close()
		if err != nil {
			return fmt.Errorf("%q: %v", name, err)
		}
	}
	return zw.Close()
}
