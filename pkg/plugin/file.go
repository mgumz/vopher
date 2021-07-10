package plugin

import (
	"bufio"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	neturl "net/url"
	"os"
	"path"
	"strings"
	"unicode/utf8"

	"github.com/mgumz/vopher/pkg/archive"
	"github.com/mgumz/vopher/pkg/utils"
)

const (
	ByteOrderMark = '\uFEFF'
)

// Parse parses a vopher-plugin file. The file format is pretty simple:
// - each plugin is stated on a line
// - empty lines or lines starting with a '#' are ignored
// - the main piece is an URL to a plugin
// - optional: a short-name for the plugin identified by the URL
// - optional: several options in terms what to do when vopher has fetched
//   the plugin (stripping paths, executing hooks etc)
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
func (plugins List) Parse(reader io.ReadCloser) error {

	defer reader.Close()

	isComment := func(fields []string) ([]string, bool) {
		return fields, (len(fields) == 0 || strings.HasPrefix(fields[0], "#"))
	}
	// the first fields is eventually the plugin name, the
	// 2nd field is then the URL
	eventualName := func(fields []string) (string, []string) {
		if len(fields) > 1 && !strings.Contains(fields[0], "://") {
			return fields[0], fields[1:]
		}
		return "", fields
	}
	// strip away .zip (or other archive-formats)
	cleanName := func(name string) string {
		if ok, l := archive.IsSupportedArchive(name); ok {
			return name[:len(name)-l]
		}
		return name
	}
	stripBom := func(in string) string {
		if !strings.ContainsRune(in, ByteOrderMark) {
			return in
		}
		return in[utf8.RuneLen(ByteOrderMark):]
	}
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
			return fmt.Errorf("existing plugin %q on line %d", name, ln)
		}

		plugin := Plugin{
			Name: name,
			URL:  url,
			Opts: Opts{StripDir: DefaultStrip},
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
	file, err := os.Open(name)
	if err != nil {
		return err
	}
	return plugins.Parse(file)
}

func (plugins List) ParseRemoteFile(url string) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	return plugins.Parse(resp.Body)
}

func (plugins List) ParseLine(line string) error {
	r := ioutil.NopCloser(strings.NewReader(line))
	return plugins.Parse(r)
}
