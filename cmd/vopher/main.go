package main

// *vopher* - acquire and manage vim-plugins by the power of the gopher.
//
// Idea: instead of having python/ruby/curl/wget/fetch/git installed
// for a vim-plugin-manager to fetch the plugins, I just want one binary
// which does it all.
//
// plugins: http://vimawesome.com/

import (
	"encoding/json"
	"flag"
	"log"
	"os"
	"strings"

	"github.com/mgumz/vopher/pkg/action"
	"github.com/mgumz/vopher/pkg/plugin"
	"github.com/mgumz/vopher/pkg/ui"
	"github.com/mgumz/vopher/pkg/utils"
)

var allowedActions = []string{
	"u",
	"up",
	"update",
	"dj",
	"dump-json",
	"fu",
	"fupdate",
	"fast-update",
	"fast",
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
	"version",
	"v",
	"archives",
	"list-archives",
	"la",
	"vim-packs", "vp", "nvim-packs", "nvp",
}

func main() {

	log.SetPrefix("vopher.")
	cli := struct {
		action  string
		force   bool
		dry     bool
		all     bool
		from    string
		dir     string
		ui      string
		filter  utils.StringList
		version bool
	}{
		action: "status",
		from:   "vopher.list",
		dir:    "./pack/vopher/start", // vim8 default "package" folder
	}

	flag.BoolVar(&cli.force, "force", cli.force, "force certain actions [prune, clean]")
	flag.BoolVar(&cli.dry, "dry", cli.dry, "dry-run, show what would happen [prune, clean]")
	flag.BoolVar(&cli.all, "all", cli.force, "don't keep <plugin>.zip around [prune]")
	flag.BoolVar(&cli.version, "v", cli.version, "show version")
	flag.StringVar(&cli.from, "f", cli.from, "path|url to list of plugins")
	flag.StringVar(&cli.dir, "dir", cli.dir, "directory to extract the plugins to")
	flag.StringVar(&cli.ui, "ui", cli.ui, "ui mode ('simple' or 'oneline', works with `update` action)")
	flag.Var(&cli.filter, "filter", "operate on given plugins only; matches substrings, can be given multiple times")

	flag.Usage = usage
	flag.Parse()

	if cli.version {
		printVersion()
		return
	}

	if len(flag.Args()) > 0 {
		cli.action = flag.Args()[0]
	}

	if utils.PrefixInStringSlice(allowedActions, cli.action) == -1 {
		log.Fatal("error: unknown action")
	}

	switch cli.action {
	case "sample":
		action.Sample()
		return
	case "version", "v":
		printVersion()
		return
	case "archives", "la", "list-archives":
		action.ListArchives()
		return
	}

	if cli.action == "search" && len(flag.Args()) < 2 {
		log.Fatal("error: missing arguments for 'search'")
	}
	if cli.action == "fetch" && len(flag.Args()) < 2 {
		log.Fatal("error: missing arguments for 'fetch'")
	}
	if cli.dir == "" {
		log.Fatal("error: empty -dir")
	}

	cli.ui = defaultUI(cli.ui, cli.action)
	ui := ui.NewUI(cli.ui)

	path, err := utils.ExpandPath(cli.dir)
	if err != nil {
		log.Fatalf("error: expanding %q failed while obtaining current user?? %s", cli.dir, err)
	}
	cli.dir = path

	switch cli.action {
	case "fupdate", "fu", "fast-update", "fup":
		parser := localOrRemoteParser(cli.from)
		plugins := mustReadPlugins(cli.from, parser, cli.filter)
		opts := action.ActUpdateOpts{Dir: cli.dir, Force: cli.force, DryRun: cli.dry}
		action.FastUpdate(plugins, ui, &opts)
	case "update", "u", "up":
		parser := localOrRemoteParser(cli.from)
		plugins := mustReadPlugins(cli.from, parser, cli.filter)
		opts := action.ActUpdateOpts{Dir: cli.dir, Force: cli.force, DryRun: cli.dry}
		action.Update(plugins, ui, &opts)
	case "fetch", "f":
		from := strings.Join(flag.Args()[1:], " ")
		plugins := mustReadPlugins(from, plugin.List.ParseLine, []string{})
		opts := action.ActUpdateOpts{Dir: cli.dir, Force: cli.force, DryRun: cli.dry}
		action.Update(plugins, ui, &opts)
	case "check", "ch":
		parser := localOrRemoteParser(cli.from)
		plugins := mustReadPlugins(cli.from, parser, cli.filter)
		action.Check(plugins, cli.dir, ui)
	case "clean", "cl":
		parser := localOrRemoteParser(cli.from)
		plugins := mustReadPlugins(cli.from, parser, cli.filter)
		action.Clean(plugins, cli.dir, cli.force)
	case "prune", "p", "pr":
		parser := localOrRemoteParser(cli.from)
		plugins := mustReadPlugins(cli.from, parser, cli.filter)
		action.Prune(plugins, cli.dir, cli.force, cli.all)
	case "status", "st":
		parser := localOrRemoteParser(cli.from)
		plugins := mayReadPlugins(cli.from, parser, cli.filter)
		action.Status(plugins, cli.dir)
	case "search", "se":
		action.Search(flag.Args()[1:]...)
	case "ping", "pong":
		action.PingPong(ui)
	case "vim-packs", "vp", "nvim-packs", "nvp":
		parser := localOrRemoteParser(cli.from)
		plugins := mustReadPlugins(cli.from, parser, cli.filter)
		action.VimPacks(plugins)
	// more of an internal, debug kind of action, so far, not documented by
	// intention
	case "dump-json", "dj":
		parser := localOrRemoteParser(cli.from)
		plugins := mustReadPlugins(cli.from, parser, cli.filter)
		enc := json.NewEncoder(os.Stdout)
		enc.Encode(plugins)
	}
}

func defaultUI(ui, action string) string {
	if ui != "" {
		return ui
	}
	switch action {
	case "update", "u", "up",
		"fetch", "f",
		"fupdate", "fu":
		return "oneline"
	default:
		return "simple"
	}
}

func mayReadPlugins(path string, parse plugin.Parser, filter utils.StringList) plugin.List {
	plugins := make(plugin.List)
	_ = parse(plugins, path)
	plugins = plugins.Filter(filter)
	return plugins
}

func mustReadPlugins(resource string, parse plugin.Parser, filter utils.StringList) plugin.List {
	plugins := make(plugin.List)
	if err := parse(plugins, resource); err != nil {
		log.Fatal(err)
	}
	plugins = plugins.Filter(filter)

	if len(plugins) == 0 {
		log.Fatalf("no plugins in %q", resource)
	}
	return plugins
}

func localOrRemoteParser(from string) plugin.Parser {
	if strings.HasPrefix(from, "http://") {
		return plugin.List.ParseRemoteFile
	} else if strings.HasPrefix(from, "https://") {
		return plugin.List.ParseRemoteFile
	}
	return plugin.List.ParseFile
}
