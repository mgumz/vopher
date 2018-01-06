package main

import (
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	neturl "net/url"
	"runtime"
	"strconv"
	"strings"
)

const (

	// most plugins are fetched from github. the github zip-files
	// put the files into a subfolder like this:
	//   vim-plugin/doc/plugin.txt
	//   vim-plugin/README.txt
	//
	DEFAULT_STRIP = 1
)

type Plugin struct {
	name    string
	ext     string
	url     *neturl.URL
	opts    PluginOpts
	archive PluginArchive // used to extract/view content of plugin
}

type PluginOpts struct {
	stripDir   int    // strip n dir-parts from archive-entries
	postUpdate string // execute after 'update'-action
	sha1       string
}

func (pl *Plugin) String() string {
	return fmt.Sprintf("Plugin{%q, %q, strip=%d}",
		pl.name, pl.url.String(), pl.opts.stripDir)
}

func (p *Plugin) optionsFromFields(fields []string) error {

	postUpdateOS := "postupdate." + runtime.GOOS + "="

	for _, field := range fields {
		if strings.HasPrefix(field, "strip=") {
			strip, err := strconv.ParseUint(field[len("strip="):], 10, 8)
			if err == nil {
				p.opts.stripDir = int(strip)
			} else {
				return fmt.Errorf("strange 'strip' field")
			}
		} else if strings.HasPrefix(field, "postupdate=") && p.opts.postUpdate == "" {
			p.opts.postUpdate = field[len("postupdate="):]
		} else if strings.HasPrefix(field, postUpdateOS) {
			p.opts.postUpdate = field[len(postUpdateOS):]
		} else if strings.HasPrefix(field, "sha1=") {
			p.opts.sha1 = strings.ToLower(field[len("sha1="):])
		}
	}

	if p.opts.postUpdate != "" {
		decoded, err := neturl.QueryUnescape(p.opts.postUpdate)
		if err != nil {
			return err
		}
		p.opts.postUpdate = decoded
	}

	if p.opts.sha1 != "" && len(p.opts.sha1) != hex.EncodedLen(sha1.Size) {
		return fmt.Errorf("'sha1' field does not match size of a sha1")
	}

	return nil
}
