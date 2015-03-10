package main

// search for vim-plugins. right now i opted to use vimawesome.com: json-api.
// a valid alternative would be to use "the source" on
//   http://www.vim.org/scripts/script_search_results.php
// con: html-scraping

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
)

const SEARCH_URL = "http://vimawesome.com/api/plugins"

func act_search(args ...string) {
	buf := bytes.NewBuffer(nil)
	for i := range args {
		// TODO: filter out special arguments such as 'page=xyz'
		buf.WriteString(args[i])
		buf.WriteByte(' ')
	}

	search_values := url.Values{}
	search_values.Add("query", buf.String())

	search_url, _ := url.Parse(SEARCH_URL)
	search_url.RawQuery = search_values.Encode()

	query, err := http.Get(search_url.String())
	if err != nil {
		log.Fatal(err)
	}
	defer query.Body.Close()

	// TODO: handle status code

	vimawesome, err := _parse_vimawesome(query.Body)
	if err != nil {
		fmt.Println("found nothing")
		return
	}

	for _, plugin := range vimawesome.Plugin {
		fmt.Println()
		// TODO: think about using http://godoc.org/github.com/kr/text
		// to improve rendering of search results
		fmt.Println(plugin.GithubUsage, plugin.Name, plugin.Desc)
		if plugin.VimUrl != "" {
			fmt.Println("      vim:", plugin.VimUrl)
		}
		if plugin.GithubUrl != "" {
			fmt.Println("   github:", plugin.GithubUrl)
		}
	}
	if len(vimawesome.Plugin) > 0 {
		search_url.Path = ""
		search_url.RawQuery = ""
		fmt.Println()
		fmt.Println("more plugins at", search_url.String())
	}
}

// {
//    "total_results": 293,
//    "results_per_page": 20,
//    "total_pages": 15,
//    "plugins": [
//       {
//         "author": "Tim Pope",
//         "category": "integration",
//         "created_at": 1255050589,
//         "github_author": "Tim Pope"},
//         "github_bundles": 10460,
//         "github_homepage": "http://www.vim.org/scripts/script.php?script_id=2975",
//         "github_owner": "tpope",
//         "github_readme_filename": "README.markdown",
//         "github_repo_id": "331603",
//         "github_repo_name": "vim-fugitive",
//         "github_short_desc": "fugitive.vim: a Git wrapper so awesome, it should be illegal",
//         "github_stars": 4638,
//         "github_url": "https://github.com/tpope/vim-fugitive",
//         "github_vim_scripts_bundles": 173,
//         "github_vim_scripts_repo_name": "fugitive.vim",
//         "github_vim_scripts_stars": 8,
//         "keywords": "a awesome, be fugitive.vim fugitive.vim: git illegal it pope should so tim wrapper",
//         "name": "fugitive.vim",
//         "normalized_name": "fugitive",
//         "plugin_manager_users": 10633,
//         "short_desc": "fugitive.vim: a Git wrapper so awesome, it should be illegal",
//         "slug": "fugitive-vim",
//         "tags": ["git"],
//         "updated_at": 1409673934,
//         "vimorg_author": "Tim Pope",
//         "vimorg_downloads": 10209,
//         "vimorg_id": "2975",
//         "vimorg_name": "fugitive.vim",
//         "vimorg_num_raters": 567,
//         "vimorg_rating": 1985,
//         "vimorg_short_desc": "A Git wrapper so awesome, it should be illegal",
//         "vimorg_type": "utility",
//         "vimorg_url": "http://www.vim.org/scripts/script.php?script_id=2975",

type _VimAwesome struct {
	Plugin []struct {
		Rating      int    `json:"vimorg_rating"`
		Name        string `json:"name"`
		Desc        string `json:"short_desc"`
		VimUrl      string `json:"vimorg_url"`
		GithubUrl   string `json:"github_url"`
		GithubUsage int    `json:"github_bundles"`
	} `json:"plugins"`
}

func _parse_vimawesome(r io.Reader) (*_VimAwesome, error) {
	var vimawesome _VimAwesome
	jsondec := json.NewDecoder(r)
	err := jsondec.Decode(&vimawesome)
	if err != nil {
		return nil, err
	}
	return &vimawesome, nil
}
