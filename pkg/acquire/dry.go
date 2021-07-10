package acquire

import (
	"bytes"
	"fmt"

	"github.com/mgumz/vopher/pkg/http"
	"github.com/mgumz/vopher/pkg/vopher"
)

// DryAcquire downloads 'url' and tries to parse the zip-file. prints out
// the files inside the zip while applying 'skip_dirs'.
//
// TODO: dryAcquire is not nice with ui 'oneline' (or future 'curses' based
// ones)
//
func DryAcquire(base, url string, archive vopher.Archive, skipDirs int, checkSha1 string) ([]string, error) {

	buffer := bytes.NewBuffer(nil)
	if err := http.Get(buffer, url, checkSha1); err != nil {
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
