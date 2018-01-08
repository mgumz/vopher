package main

// *vopher* - acquire and manage vim-plugins by the power of the gopher
//
// idea: instead of having python/ruby/curl/wget/fetch/git installed
// for a vim-plugin-manager to fetch the plugins i just want one binary
// which does it all.
//
// plugins: http://vimawesome.com/

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strings"
	"time"
)

var allowedActions = []string{
	"u",
	"up",
	"update",
	"f",
	"fetch",
	"clean",
	"c",
	"check",
	"ping",
	"prune",
	"sample",
	"st",
	"status",
	"search",
}

func usage() {

	fmt.Fprintln(os.Stderr, `vopher - acquire vim plugins the gopher-way

Usage: vopher [flags] <action>

Flags:
`)
	flag.PrintDefaults()
	fmt.Fprintln(os.Stderr, `
Actions:

  update - acquires the given plugins from '-f <list>'
  fetch  - fetch a remote archive and extract it. the arguments are like fields
           in a vopher.list file
  search - searches http://vimawesome.com/ to list some plugins. Anything
           after this is considered the query
  check  - checks plugins from '-f <list>' for newer versions
  clean  - removes given plugins from the '-f <list>'
           * use '-force' to delete plugins.
  prune  - removes all entries from -dir <folder> which are not referenced in
           '-f <list>'.
           * use '-force' to delete plugins.
           * use '-all=true' to delete <plugin>.zip files.
  status - lists plugins in '-dir <folder>' and marks them accordingly
           * 'v' means vopher is tracking the plugin in your '-f <list>'
           * 'm' means vopher is tracking the plugin and it's missing. You can
             fetch it with the 'update' action.
           * no mark means that the plugin is not tracked by vopher
  sample - prints a sample vopher.list to stdout`)
}

func main() {

	log.SetPrefix("vopher.")
	cli := struct {
		action    string
		force     bool
		dry       bool
		all       bool
		file      string
		dir       string
		ui        string
		filter    stringList
		supported bool
		version   bool
	}{
		action: "status",
		file:   "vopher.list",
		dir:    "./pack/vopher/start", // vim8 default "package" folder
	}

	flag.BoolVar(&cli.force, "force", cli.force, "force certain actions [prune, clean]")
	flag.BoolVar(&cli.dry, "dry", cli.dry, "dry-run, show what would happen [prune, clean]")
	flag.BoolVar(&cli.all, "all", cli.force, "don't keep <plugin>.zip around [prune]")
	flag.BoolVar(&cli.supported, "list-supported-archives", false, "list all supported archive types")
	flag.BoolVar(&cli.version, "v", cli.version, "show version")
	flag.StringVar(&cli.file, "f", cli.file, "path to list of plugins")
	flag.StringVar(&cli.dir, "dir", cli.dir, "directory to extract the plugins to")
	flag.StringVar(&cli.ui, "ui", cli.ui, "ui mode ('simple' or 'oneline', works with `update` action)")
	flag.Var(&cli.filter, "filter", "operate on given plugins only; matches substrings, can be given multiple times")

	flag.Usage = usage
	flag.Parse()

	if cli.version {
		printVersion()
		return
	}

	if cli.supported {
		for _, suf := range supportedArchives {
			fmt.Println(suf)
		}
		return
	}

	if len(flag.Args()) > 0 {
		cli.action = flag.Args()[0]
	}

	if prefixInStringSlice(allowedActions, cli.action) == -1 {
		log.Fatal("error: unknown action")
	}

	if cli.action == "sample" {
		actSample()
		return
	} else if cli.action == "search" && len(flag.Args()) < 2 {
		log.Fatal("error: missing arguments for 'search'")
	} else if cli.action == "fetch" && len(flag.Args()) < 2 {
		log.Fatal("error: missing arguments for 'fetch'")
	}

	cli.ui = defaultUI(cli.ui, cli.action)
	var ui JobUI = generateUI(cli.ui)

	if cli.dir == "" {
		log.Fatal("error: empty -dir")
	}

	path, err := expandPath(cli.dir)
	if err != nil {
		log.Fatalf("error: expanding %q failed while obtaining current user?? %s", cli.dir, err)
	}
	cli.dir = path

	switch cli.action {
	case "update", "u", "up":
		plugins := mustReadPlugins(cli.file, cli.filter)
		opts := ActUpdateOpts{dir: cli.dir, force: cli.force, dryRun: cli.dry}
		actUpdate(plugins, ui, &opts)
	case "fetch", "f":
		plugins := fetchSinglePlugin(strings.Join(flag.Args()[1:], " "))
		opts := ActUpdateOpts{dir: cli.dir, force: cli.force, dryRun: cli.dry}
		actUpdate(plugins, ui, &opts)
	case "check", "ch":
		plugins := mustReadPlugins(cli.file, cli.filter)
		actCheck(plugins, cli.dir, ui)
	case "clean", "cl":
		plugins := mustReadPlugins(cli.file, cli.filter)
		actClean(plugins, cli.dir, cli.force)
	case "prune", "p", "pr":
		plugins := mustReadPlugins(cli.file, cli.filter)
		actPrune(plugins, cli.dir, cli.force, cli.all)
	case "status", "st":
		plugins := mayReadPlugins(cli.file, cli.filter)
		actStatus(plugins, cli.dir)
	case "search", "se":
		actSearch(flag.Args()[1:]...)
	case "ping", "pong":
		actPingPong(ui)
	}
}

func defaultUI(ui, action string) string {
	if ui != "" {
		return ui
	}
	switch action {
	case "update", "u", "up",
		"fetch", "f":
		return "oneline"
	default:
		return "simple"
	}
}

func generateUI(ui string) JobUI {
	switch ui {
	case "oneline":
		return &UIOneLine{
			pt:       newProgressTicker(0),
			prefix:   "vopher",
			duration: 25 * time.Millisecond,
		}
	case "simple":
		return &UISimple{jobs: make(map[string]Runtime)}
	case "quiet":
		return &UIQuiet{}
	}
	return nil
}

func mayReadPlugins(path string, filter stringList) PluginList {
	plugins := make(PluginList)
	plugins.ParseFile(path)
	plugins = plugins.filter(filter)
	return plugins
}

func mustReadPlugins(path string, filter stringList) PluginList {
	plugins := make(PluginList)
	if err := plugins.ParseFile(path); err != nil {
		log.Fatal(err)
	}
	plugins = plugins.filter(filter)

	if len(plugins) == 0 {
		log.Fatalf("empty plugin-file %q", path)
	}
	return plugins
}

func fetchSinglePlugin(url string) PluginList {

	r := ioutil.NopCloser(strings.NewReader(url))
	plugins := make(PluginList)
	if err := plugins.Parse(r); err != nil {
		log.Fatal(err)
	}
	if len(plugins) == 0 {
		log.Fatalf("not a valid plugin %q", url)
	}
	return plugins
}
