package acquire

import (
	"bytes"
	"fmt"
	"os"

	"github.com/mgumz/vopher/pkg/http"
	"github.com/mgumz/vopher/pkg/vopher"
)

// DryAcquire downloads src.URL (offline: reads src.Path) and tries to parse
// the zip-file. prints out the files inside the zip while applying
// 'skip_dirs'.
//
// TODO: dryAcquire is not nice with ui 'oneline' (or future 'curses' based
// ones)
func DryAcquire(src Source, archive vopher.Archive, skipDirs int) ([]string, error) {

	buffer := bytes.NewBuffer(nil)
	if src.Offline {
		data, err := os.ReadFile(src.Path)
		if err != nil {
			return nil, fmt.Errorf("%q not available offline: %s", src.URL, err)
		}
		buffer.Write(data)
	} else if err := http.Get(buffer, src.URL, src.SHA1); err != nil {
		return nil, err
	}
	defer buffer.Reset()

	br := bytes.NewReader(buffer.Bytes())
	entries, err := archive.Entries(br, skipDirs)

	if err != nil {
		return nil, fmt.Errorf("gettting contents: %v", err)
	}
	if len(entries) == 0 {
		return nil, fmt.Errorf("empty archive for %q", src.URL)
	}

	return entries, nil
}
