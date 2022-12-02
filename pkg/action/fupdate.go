package action

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/mgumz/vopher/pkg/plugin"
	"github.com/mgumz/vopher/pkg/ui"
)

const backupTag = ".2006-01-02T03-04-05Z"

func FastUpdate(plugins plugin.List, ui ui.UI, opts *ActUpdateOpts) {

	ui.Start()

	fn := opts.Dir

	if fi, err := os.Stat(fn); err == nil && fi.IsDir() {
		fnBackup := fn + time.Now().UTC().Format(backupTag)
		ui.Print("fast-update", fmt.Sprintf("backup plugins %s", fnBackup))
		err := os.Rename(fn, fnBackup)
		if err != nil {
			log.Printf("error: can't rename %q to %q: %s",
				fn, fnBackup, err)
			return
		}
		ui.Print("fast-update", "done.")
	}

	Update(plugins, ui, opts)
}
