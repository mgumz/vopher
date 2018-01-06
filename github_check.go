package main

import (
	"bytes"
	"fmt"
	"sync"
)

func (gh Github) checkPlugin(plugin *Plugin, base string) string {

	name, head := gh.getRepository(plugin.url)
	altHead := gh.guessCommitByZIP(plugin.name, base)

	u := *plugin.url
	u.Path = name

	commits := []struct {
		parts   []string
		feed    *GithubFeed
		section string
	}{
		{[]string{"commits", "master"}, nil, "\n - master commits:\n"},
		{[]string{"commits", head}, nil, "\n - commits:\n"},
		{[]string{"tags"}, nil, "\n - tags:\n"},
	}

	var wg sync.WaitGroup
	wg.Add(len(commits))
	for i := range commits {
		if i == 0 && head == "master" { // don't check 'master' two times
			wg.Done()
			continue
		}
		go func(j int) {
			commits[j].feed = gh.getCommits(&u, commits[j].parts...)
			wg.Done()
		}(i)
	}
	wg.Wait()

	buf := bytes.NewBuffer(nil)

	fmt.Fprintf(buf, "\n\n## %s - %s\n", plugin.name, plugin.url)

	prefix := " "
	for i := range commits {
		c := &commits[i]
		if c.feed == nil || len(c.feed.Entry) <= 0 {
			continue
		}

		fmt.Fprintln(buf, c.section)
		for _, entry := range c.feed.Entry {
			prefix = " "
			if entry.ID == head || entry.ID == altHead {
				prefix = "*"
			}
			fmt.Fprintf(buf, "  %s%.10s %s %s\n", prefix, entry.ID, entry.Updated, entry.Title)
		}
	}

	return buf.String()
}
