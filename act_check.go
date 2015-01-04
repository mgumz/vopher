package main

import (
	"archive/zip"
	"bytes"
	"encoding/xml"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"path"
	"path/filepath"
	"strings"
	"sync"
)

func act_check(plugins PluginList, base string, ui JobUi) {

	for _, plugin := range plugins {
		switch plugin.url.Host {
		case "github.com":

			name, head := _gh_get_repository(plugin.url)
			alt_head := _gh_guess_commit_by_zip(plugin.name, base)

			u := *plugin.url
			u.Path = name

			commits := []struct {
				parts   []string
				atom    *_GhAtom
				section string
			}{
				{[]string{"commits", "master"}, nil, "\n - master commits:\n"},
				{[]string{"commits", head}, nil, "\n - commits:\n"},
				{[]string{"tags"}, nil, "\n - tags:\n"},
			}

			wg := sync.WaitGroup{}
			wg.Add(len(commits))
			for i := range commits {
				if head == "master" {
					wg.Done()
					continue
				}
				go func(j int) {
					commits[j].atom = _gh_get_commits(&u, commits[j].parts...)
					wg.Done()
				}(i)
			}
			wg.Wait()

			buf := bytes.NewBuffer(nil)

			fmt.Fprintf(buf, "\n\n## %s - %s\n", plugin.name, plugin.url)

			prefix := " "
			for i := range commits {
				c := &commits[i]
				if c.atom == nil || len(c.atom.Entry) <= 0 {
					continue
				}

				fmt.Fprintln(buf, c.section)
				for _, entry := range c.atom.Entry {
					prefix = " "
					if entry.Id == head || entry.Id == alt_head {
						prefix = "*"
					}
					fmt.Fprintf(buf, "  %s%.10s %s %s\n", prefix, entry.Id, entry.Updated, entry.Title)
				}
			}

			ui.Print(plugin.name, buf.String())
		}
	}
}

// the repo-name is usually the first 2 parts of the
// repo.Path:
//    github.com/username/reponame
//    github.com/username/reponame/archive/master.zip
//    github.com/username/reponame#v2.1
//
func _gh_get_repository(remote *url.URL) (name, head string) {
	name = remote.Path
	if idx := index_byte_n(remote.Path, '/', 3); idx > 0 {
		name = name[:idx]
	}

	// TODO: other means to detect the current used 'head'
	//
	head = "master"
	if remote.Fragment != "" {
		head = remote.Fragment
	} else if strings.HasSuffix(remote.Path, ".zip") {
		head = path.Base(remote.Path)
		head = head[:len(head)-4]
	}

	return
}

// minimal atom-parser sufficient to extract only what we need
// from github
type _GhAtom struct {
	XMLName xml.Name `xml:"feed"`
	Entry   []struct {
		Title   string `xml:"title"`
		Updated string `xml:"updated"`
		Id      string `xml:"id"`
	} `xml:"entry"`
}

func _gh_get_commits(repo *url.URL, parts ...string) *_GhAtom {

	atom_url := *repo
	atom_url.Path = path.Join(atom_url.Path, path.Join(parts...)) + ".atom"

	resp, err := http.Get(atom_url.String())
	if err != nil {
		log.Println("error", atom_url, err)
		return nil
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		log.Println("error", atom_url, err)
		return nil
	}

	xmldec := xml.NewDecoder(resp.Body)
	gh := _GhAtom{}

	if err = xmldec.Decode(&gh); err != nil {
		log.Println("error", atom_url, err)
		return nil
	}

	if len(gh.Entry) == 0 {
		return nil
	}

	for i, entry := range gh.Entry {
		idx := strings.LastIndex(entry.Id, "/")
		id := entry.Id[idx+1:]
		title := strings.TrimSpace(entry.Title)
		gh.Entry[i].Id = id
		gh.Entry[i].Title = title
	}

	return &gh
}

func _gh_guess_commit_by_zip(name, base string) string {

	path := filepath.Join(base, name+".zip")
	zfile, err := zip.OpenReader(path)
	if err != nil {
		log.Println(path)
		return ""
	}
	defer zfile.Close()

	if len(zfile.Comment) == 40 {
		return zfile.Comment
	}
	return ""
}
