package update

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"

	"github.com/mgumz/vopher/pkg/acquire"
	"github.com/mgumz/vopher/pkg/plugin"
	"github.com/mgumz/vopher/pkg/ui"
	"github.com/mgumz/vopher/pkg/utils"
)

func AcquireAndPostupdate(dir string, dryRun bool, plugin *plugin.Plugin, ui ui.UI) {

	defer ui.JobDone(dir)

	var (
		err     error
		path    string
		out     []byte
		entries []string

		url      = plugin.URL.String()
		stripDir = plugin.Opts.StripDir
		sha1     = plugin.Opts.SHA1
	)

	if dryRun {
		entries, err = acquire.DryAcquire(dir, url, plugin.Archive, stripDir, sha1)
	} else {
		err = acquire.Acquire(dir, plugin.Ext, url, plugin.Archive, stripDir, sha1)
	}
	if err != nil {
		log.Printf("%s: %v", dir, err)
		return
	}
	if dryRun {
		ui.Print(dir, url)
		ui.Print(dir, strings.Join(entries, "\n"))
	}

	//
	// no postUpdate hook? done.
	//
	if plugin.Opts.PostUpdate == "" {
		return
	}

	//
	// handle the .postupdate hook
	//
	path, err = utils.ExpandPath(plugin.Opts.PostUpdate)
	if err != nil {
		log.Printf("%s: expanding .postupdate %q: %s", dir, plugin.Opts.PostUpdate, err)
		return
	}
	path = utils.ExpandPathEnvironment(path, dir)

	// we won't check for existing executable, we just prepare the
	// right environment and then we launch it. if it fails, the OS will tell
	// us (missing permissions, file not found etc)
	//
	// we do not set .Dir because this will cause golang to search for .Path
	// underneath it. if the user wants to call a script inside .Dir (which
	// should be a rare case anyway) then she should use $VOPHER_DIR/cmd which
	// gets expanded, see above.
	cmd := buildCmd(path, plugin.Name, plugin.Ext, dir, url)

	if dryRun {
		ui.Print(dir, fmt.Sprintf("# postupdate: %q (env: %v)\n", cmd.Path, cmd.Env))
		return
	}

	out, err = cmd.CombinedOutput()
	if err != nil {
		log.Printf("%s: error running .postupdate %q: %s", dir, path, err)
		log.Printf("%s: %s", dir, string(out))
		return
	}

	if len(out) == 0 {
		return
	}

	ui.Print(dir, string(out))
}

func buildCmd(path, name, ext, dir, url string) exec.Cmd {

	return exec.Cmd{
		Path: path,
		Env: append(os.Environ(),
			"VOPHER_NAME="+name,
			"VOPHER_ARCHIVE="+name+ext,
			"VOPHER_DIR="+dir,
			"VOPHER_URL="+url),
	}
}
