package acquire

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/mgumz/vopher/pkg/http"
	"github.com/mgumz/vopher/pkg/vopher"
)

// Acquire fetches 'url' to 'archivePath' and extracts it into 'base'. skips
// 'skipDirs' leading directories in filenames in zip while extracting
// the contents.
func Acquire(base, archivePath, url string, archive vopher.Archive, skipDirs int, checkSha1 string) error {

	const dirPerms = 0700
	err := os.MkdirAll(base, dirPerms)

	if err != nil {
		return fmt.Errorf("mkdir %q: %s", base, err)
	}

	cacheDir := filepath.Dir(archivePath)
	if err = os.MkdirAll(cacheDir, dirPerms); err != nil {
		return fmt.Errorf("mkdir %q: %s", cacheDir, err)
	}

	ts := time.Now().UTC().Format("-2006-01-02T03-04-05Z")
	tmpName := archivePath + ts
	file, err := os.Create(tmpName) // #nosec G304
	if err != nil {
		return err
	}

	if err = http.Get(file, url, checkSha1); err != nil {
		_ = file.Close()
		_ = os.Remove(tmpName)
		return err
	}

	_ = file.Sync()
	_, _ = file.Seek(0, 0)

	err = archive.Extract(base, file, skipDirs)
	_ = file.Close()
	if err == nil {
		err = os.Rename(tmpName, archivePath)
	}
	return err
}
