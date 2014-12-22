package main

// idea: instead of having python/ruby/curl/wget/fetch/git installed
// for a vim-plugin-manager to fetch the plugins i just want one binary
// which does it all.
//
// plugins: http://vimawesome.com/
//
// ui-options:
//
// * https://godoc.org/github.com/jroimartin/gocui
//
//  global-progress [..............]
//  plugin1         [....]
//  plugin2         [............]
//  plugin3         [..............]
//
// cons: vertical space
//
// ui-option2:
//   <-> global progress
//  [....|.....|.....|....|....|....]
//   ^
//   | plugin-progress via _-=#░█▓▒░█
//   v
//
// cons: horizontal space
//        plugin-name fehlt

import (
	"archive/zip"
	"flag"
	"io"
	"log"
	"os"
	"path"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

type JobWait struct {
	sync.WaitGroup
	*ProgressTicker
}

func (jw *JobWait) Done() { jw.ProgressTicker.WriteCounter += 1; jw.WaitGroup.Done() }
func (jw *JobWait) Wait() { jw.WaitGroup.Wait(); jw.ProgressTicker.MaxOut() }

func main() {
	force := flag.Bool("force", false, "force download of existing plugins")
	depfile := flag.String("pf", "plugins.lst", "plugins.lst")
	base := flag.String("plugins", ".", "path to extract the plugins to")
	flag.Parse()

	plugins, err := ScanPluginFile(*depfile)
	if err != nil {
		log.Fatal(err)
	}
	if len(plugins) == 0 {
		log.Fatal("empty plugin-file")
	}

	wg := JobWait{ProgressTicker: NewProgressTicker(int64(len(plugins)))}
	go wg.Start("vopher", 25*time.Millisecond)

	for _, plugin := range plugins {

		plugin_folder := filepath.Join(*base, plugin.name)

		_, err := os.Stat(plugin_folder)
		if err == nil { // plugin_folder exists
			if !*force {
				continue
			}
		}

		if !strings.HasSuffix(plugin.url.Path, ".zip") {
			switch plugin.url.Host {
			case "github.com":
				remote_zip := first_not_empty(plugin.url.Fragment, "master") + ".zip"
				plugin.url.Path = path.Join(plugin.url.Path, "archive", remote_zip)
			default:
				ext, err := httpdetect_ftype(plugin.url.String())
				if err != nil {
					log.Printf("error: %q: %s", plugin.url, err)
					continue
				}
				if ext != ".zip" {
					log.Printf("error: %q: not a zip", plugin.url)
					continue
				}
			}
		}

		wg.Add(1)
		go fetch_and_extract(&wg, plugin_folder, plugin.url.String(), plugin.strip_dir)
	}
	wg.Wait()
	wg.Stop()
}

func fetch_and_extract(wg *JobWait, base, url string, skip_dirs int) {
	defer wg.Done()

	if err := os.MkdirAll(base, 0777); err != nil {
		log.Println("mkdir", base, err)
		return
	}

	name := base + ".zip"
	if err := httpget(name, url); err != nil {
		log.Println(url, err)
		return
	}
	zfile, err := zip.OpenReader(name)
	if err != nil {
		log.Println(name, err)
		return
	}
	defer zfile.Close()
	for _, f := range zfile.File {
		idx := index_byte_n(f.Name, '/', skip_dirs)

		oname := f.Name[idx+1:]

		// root-directory
		//   pname/      <- root-directory
		//   pname/a.vim
		if oname == "" {
			continue
		}

		oname = filepath.Join(base, filepath.Clean(oname))

		if f.FileInfo().IsDir() {
			os.MkdirAll(oname, 0777)
			continue
		}

		// TODO: call only if needed
		os.MkdirAll(filepath.Dir(oname), 0777)

		zreader, err := f.Open()
		if err != nil {
			log.Println(oname, err)
		}
		ofile, err := os.Create(oname)
		if err != nil {
			log.Println(oname, err)
		}
		_, err = io.Copy(ofile, zreader)
		if err != nil {
			log.Println(oname, err)
		}

		ofile.Close()
		zreader.Close()
	}
}
