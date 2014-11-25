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
	"net/http"
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
		if !strings.HasSuffix(plugin.url.Path, ".zip") {
			switch plugin.url.Host {
			case "github.com":
				remote_zip := first_not_empty(plugin.url.Fragment, "master") + ".zip"
				plugin.url.Path = path.Join(plugin.url.Path, "archive", remote_zip)
			}
		}
		wg.Add(1)
		go fetch_and_extract(&wg, plugin_folder, plugin.url.String())
	}
	wg.Wait()
	wg.Stop()
}

func fetch_and_extract(wg *JobWait, base, url string) {
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
		oname := f.Name[strings.IndexByte(f.Name, '/')+1:]
		oname = filepath.Join(base, filepath.Clean(oname))
		//fmt.Println(f.Name, "=>", oname)
		if f.FileInfo().IsDir() {
			os.MkdirAll(oname, 0777)
			continue
		}
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

func httpget(out, url string) (err error) {

	var file *os.File
	var resp *http.Response

	if file, err = os.Create(out); err != nil {
		return err
	}
	defer file.Close()

	if resp, err = http.Get(url); err != nil {
		return err
	}
	defer resp.Body.Close()

	reader := io.Reader(resp.Body)
	/*
		if resp.ContentLength > 0 {
			progress := NewProgressTicker(resp.ContentLength)
			defer progress.Stop()
			go progress.Start(out, 2*time.Millisecond)
			reader = io.TeeReader(reader, progress)
		}
	*/

	_, err = io.Copy(file, reader)
	return err
}
