package acquire

import (
	"crypto/sha1" /* #nosec */
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/mgumz/vopher/pkg/http"
	"github.com/mgumz/vopher/pkg/vopher"
)

// MetaSuffix is appended to an archive path to name the file holding the
// http.Validators of the archive
const MetaSuffix = ".meta"

// Source describes where an archive comes from and where it is kept
type Source struct {
	URL       string
	Path      string // where the archive is kept
	SHA1      string // expected checksum, optional
	Immutable bool   // URL always yields the same content, eg a git commit
	Offline   bool   // never fetch, use the kept archive only
}

// Acquire extracts the archive described by 'src' into 'base'. skips
// 'skipDirs' leading directories in filenames in zip while extracting
// the contents.
//
// The archive kept at src.Path is used without asking the server if src is
// Offline, Immutable or its SHA1 matches. Otherwise it is fetched
// conditionally: if the server reports it unchanged, the kept one is used.
func Acquire(base string, src Source, archive vopher.Archive, skipDirs int) error {

	const dirPerms = 0700
	err := os.MkdirAll(base, dirPerms)

	if err != nil {
		return fmt.Errorf("mkdir %q: %s", base, err)
	}

	cacheDir := filepath.Dir(src.Path)
	if err = os.MkdirAll(cacheDir, dirPerms); err != nil {
		return fmt.Errorf("mkdir %q: %s", cacheDir, err)
	}

	kept, err := fileSHA1(src.Path)
	switch {
	case err == nil:
		matches := src.SHA1 == "" || src.SHA1 == kept
		if matches && (src.Offline || src.Immutable || src.SHA1 != "") {
			return extractFile(base, src.Path, archive, skipDirs)
		}
		if src.Offline {
			return fmt.Errorf("sha1 of %q does not match: got %s, expected %s", src.Path, kept, src.SHA1)
		}
	case src.Offline:
		return fmt.Errorf("%q not available offline: %s", src.URL, err)
	}

	// a kept archive not matching src.SHA1 must not be confirmed by a 304
	have := http.Validators{}
	if kept != "" && src.SHA1 == "" {
		have = readMeta(src.Path)
	}

	ts := time.Now().UTC().Format("-2006-01-02T03-04-05Z")
	tmpName := src.Path + ts
	file, err := os.Create(tmpName) // #nosec G304
	if err != nil {
		return err
	}

	got, notModified, err := http.GetIfChanged(file, src.URL, src.SHA1, have)
	if err != nil || notModified {
		_ = file.Close()
		_ = os.Remove(tmpName)
		if err != nil {
			return err
		}
		return extractFile(base, src.Path, archive, skipDirs)
	}

	_ = file.Sync()
	_, _ = file.Seek(0, 0)

	err = archive.Extract(base, file, skipDirs)
	_ = file.Close()
	if err == nil {
		err = os.Rename(tmpName, src.Path)
	}
	if err == nil {
		writeMeta(src.Path, got)
	}
	return err
}

func extractFile(base, name string, archive vopher.Archive, skipDirs int) error {
	file, err := os.Open(name) // #nosec G304
	if err != nil {
		return err
	}
	defer file.Close()
	return archive.Extract(base, file, skipDirs)
}

func fileSHA1(name string) (string, error) {
	file, err := os.Open(name) // #nosec G304
	if err != nil {
		return "", err
	}
	defer file.Close()
	hasher := sha1.New() /* #nosec */
	if _, err = io.Copy(hasher, file); err != nil {
		return "", err
	}
	return hex.EncodeToString(hasher.Sum(nil)), nil
}

func readMeta(archivePath string) http.Validators {
	v := http.Validators{}
	data, err := os.ReadFile(archivePath + MetaSuffix) // #nosec G304
	if err == nil {
		_ = json.Unmarshal(data, &v)
	}
	return v
}

// writeMeta keeps v next to the archive; without validators an old meta
// file would describe another version, so it is removed
func writeMeta(archivePath string, v http.Validators) {
	name := archivePath + MetaSuffix
	if v == (http.Validators{}) {
		_ = os.Remove(name)
		return
	}
	data, err := json.Marshal(v)
	if err != nil {
		return
	}
	_ = os.WriteFile(name, data, 0600)
}
