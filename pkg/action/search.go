package action

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

// SearchURL to be used when using the `search` action
const SearchURL = "http://vimawesome.com/api/plugins"

// Search searches SearchURL for given args to find a (n)vim plugin
func Search(args ...string) {
	buf := bytes.NewBuffer(nil)
	for i := range args {
		// TODO: filter out special arguments such as 'page=xyz'
		buf.WriteString(args[i])
		buf.WriteByte(' ')
	}

	vals := url.Values{}
	vals.Add("query", buf.String())

	searchURL, _ := url.Parse(SearchURL)
	searchURL.RawQuery = vals.Encode()

	query, err := http.Get(searchURL.String())
	if err != nil {
		log.Fatal(err)
	}
	defer (func() { _ = query.Body.Close() })()

	// TODO: handle status code

	vimawesome := &_VimAwesome{}
	if err := vimawesome.parse(query.Body); err != nil {
		fmt.Println("found nothing")
		return
	}

	for _, p := range vimawesome.Plugin {
		fmt.Println()
		// TODO: think about using http://godoc.org/github.com/kr/text
		// to improve rendering of search results
		fmt.Println(p.GithubUsage, p.Name, p.Desc)
		if p.VimURL != "" {
			fmt.Println("      vim:", p.VimURL)
		}
		if p.GithubURL != "" {
			fmt.Println("   github:", p.GithubURL)
		}
	}
	if len(vimawesome.Plugin) > 0 {
		searchURL.Path = ""
		searchURL.RawQuery = ""
		fmt.Println()
		fmt.Println("more plugins at", searchURL.String())
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
		VimURL      string `json:"vimorg_url"`
		GithubURL   string `json:"github_url"`
		GithubUsage int    `json:"github_bundles"`
	} `json:"plugins"`
}

func (vimAwesome *_VimAwesome) parse(r io.Reader) error {
	jsondec := json.NewDecoder(r)
	return jsondec.Decode(vimAwesome)
}
