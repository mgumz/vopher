package plugin

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"net/http"
	neturl "net/url"
	"os"
	"path"
	"strings"

	"github.com/mgumz/vopher/pkg/utils"
)

const (
	ByteOrderMark = '\uFEFF'
)

// Parse parses a vopher-plugin file. The file format is pretty simple:
//   - each plugin is stated on a line
//   - empty lines or lines starting with a '#' are ignored
//   - the main piece is an URL to a plugin
//   - optional: a short-name for the plugin identified by the URL
//   - optional: several options in terms what to do when vopher has fetched
//     the plugin (stripping paths, executing hooks etc)
//
// Sample:
//
// # a comment starts with a '#'
// # empty lines are ignored
//
// # fetches vim-fugitive, current HEAD
// https://github.com/tpope/vim-fugitive
//
// # fetches vim-fugitive, tagged release 'v2.1'
// https://github.com/tpope/vim-fugitive#v2.1.zip
//
// # fetches vim-fugitive, commit 913fff1cea3aa1a08a360a494fa05555e59147f5
// https://github.com/tpope/vim-fugitive#913fff1cea3aa1a08a360a494fa05555e59147f5
//
// # fetches vim-fugitive, commit 913fff1cea3aa1a08a360a494fa05555e59147f5,
// # but since it's too long, just use the short-version of it
// https://github.com/tpope/vim-fugitive#913fff1c
//
// # fetches vim-fugitive, name the output-folder 'foo'
// foo https://github.com/tpope/vim-fugitive
//
// # fetches vim-fugitive, apply some options to the
// # parser/extractor
// https://github.com/tpope/vim-fugitive option1=foo option2=bar
//
// # fetches vim-fugitive, call a post-update cmd. vophers
// # sets the following environment variables:
// # VOPHER_NAME    - plugin name
// # VOPHER_ARCHIVE - plugin name.ext
// # VOPHER_DIR     - plugin folder
// # VOPHER_URL     - url of plugin
// https://github.com/tpope/vim-fugitive postupdate=/path/to/cmd
//
// # variant of postupdate: postupdate.linux=/path/to/cmd
func (plugins List) Parse(reader io.ReadCloser) error {

	defer (func() { _ = reader.Close() })()

	trimLine := func(in string) string {
		return strings.TrimSpace(stripBom(in))
	}

	scanner := bufio.NewScanner(reader)
	scanner.Split(bufio.ScanLines)

	for ln := 1; scanner.Scan(); ln++ {

		var line, name string
		var fields []string
		var skip bool
		var url *neturl.URL
		var err error

		// only first line might contain the Bom. switch to regular
		// strings.TrimSpace afterwards
		line, trimLine = trimLine(scanner.Text()), strings.TrimSpace
		if fields, skip = isComment(strings.Fields(line)); skip {
			continue
		}
		name, fields = eventualName(fields)
		if url, err = utils.ParsePluginURL(fields[0]); err != nil {
			log.Println("error:", name, ":", ln, "not an url", line)
			continue
		}
		name = utils.FirstNotEmpty(name, path.Base(url.Path))
		name = cleanName(name)
		if _, skip = plugins[name]; skip {
			log.Printf("existing plugin %q on line %d", name, ln)
			continue
		}

		plugin := Plugin{
			Name: name,
			URL:  url,
			Opts: defaultOpts,
			ln:   ln,
		}

		if len(fields) > 1 {
			fields = fields[1:]
			if err = plugin.optionsFromFields(fields); err != nil {
				errMsg := "parsing optional fields: %q, plugin %q on line %d"
				return fmt.Errorf(errMsg, err, name, ln)
			}
		}

		plugins[name] = &plugin
	}
	return nil
}

// Parser is a function signature to be used by
// different parsers
type Parser func(List, string) error

func (plugins List) ParseFile(name string) error {
	file, err := os.Open(name) // #nosec G304
	if err != nil {
		return err
	}
	return plugins.Parse(file)
}

func (plugins List) ParseRemoteFile(url string) error {
	resp, err := http.Get(url) // #nosec G107
	if err != nil {
		return err
	}
	return plugins.Parse(resp.Body)
}

func (plugins List) ParseLine(line string) error {
	r := io.NopCloser(strings.NewReader(line))
	return plugins.Parse(r)
}
