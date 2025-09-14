package acquire

import (
	"fmt"
	"os"
	"time"

	"github.com/mgumz/vopher/pkg/http"
	"github.com/mgumz/vopher/pkg/vopher"
)

// Acquire fetches 'url' and extracts it into 'base'. skips 'skipDirs'
// leading directories in filenames in zip while extracting
// the contents.
func Acquire(base, ext, url string, archive vopher.Archive, skipDirs int, checkSha1 string) error {

	const dirPerms = 0700
	err := os.MkdirAll(base, dirPerms)

	if err != nil {
		return fmt.Errorf("mkdir %q: %s", base, err)
	}

	ts := time.Now().UTC().Format("-2006-01-02T03-04-05Z")
	name, tmpName := base+ext, base+ts+ext
	file, err := os.Create(tmpName)
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
		err = os.Rename(tmpName, name)
	}
	return err
}
