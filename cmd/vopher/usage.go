package main

import (
	"flag"
	"fmt"
	"os"
)

const (
	usagePrefix string = `vopher - acquire vim plugins the gopher-way

Usage: vopher [flags] <action>

Flags:`

	usageActions string = `
Actions:

  update   - acquires the given plugins from '-f <list-file|url>'
  fupdate  - fast update - renames current -dir and fetches everything
  fetch    - fetch a remote archive and extract it. the arguments are like fields
             in a vopher.list file
  vp       - produces a vimscript full of "packadd!" for matching plugins
             (aliases: 'vim-packs', 'nvim-packs', 'nvp')
  search   - searches http://vimawesome.com/ to list some plugins. Anything
             after this is considered as "the search arguments"
  check    - checks plugins from '-f <list>' for newer versions
  clean    - removes given plugins from the '-f <list>'
             * use '-force' to delete plugins.
  prune    - removes all entries from -dir <folder> which are not referenced in
             '-f <list>'.
             * use '-force' to delete plugins.
             * use '-all=true' to delete <plugin>.zip files.
  status   - lists plugins in '-dir <folder>' and marks them accordingly
             * 'v' means vopher is tracking the plugin in your '-f <list>'
             * 'm' means vopher is tracking the plugin and it's missing. You can
               fetch it with the 'update' action.
             * no mark means that the plugin is not tracked by vopher
  sample   - prints a sample vopher.list to stdout
  version  - prints version of vopher
  archives - list supported archives`
)

func usage() {
	fmt.Fprintln(os.Stderr, usagePrefix)
	flag.PrintDefaults()
	fmt.Fprintln(os.Stderr, usageActions)
}
