package main

import (
	"fmt"
	neturl "net/url"
)

const (

	// most plugins are fetched from github. the github zip-files
	// put the files into a subfolder like this:
	//   vim-plugin/doc/plugin.txt
	//   vim-plugin/README.txt
	//
	DEFAULT_STRIP = 1

	OPT_STRIP_DIR     = 0
	OPT_POSTUPDATE    = 1
	OPT_POSTUPDATE_OS = 2
	OPT_SHA1          = 3
)

type Plugin struct {
	name string
	ext  string
	url  *neturl.URL
	opts struct {
		strip_dir  int    // strip n dir-parts from archive-entries
		postupdate string // execute after 'update'-action
		sha1       string
	}
	archive PluginArchive // used to extract/view content of plugin
}

func (pl *Plugin) String() string {
	return fmt.Sprintf("Plugin{%q, %q, strip=%d}",
		pl.name, pl.url.String(), pl.opts.strip_dir)
}
