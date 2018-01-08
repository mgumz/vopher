package main

import (
	"bytes"
	"fmt"
	"sync"
)

func (gh Github) checkPlugin(plugin *Plugin, base string) string {

	name, head := gh.getRepository(plugin.url)
	altHead := gh.guessCommitByFile(plugin.name, base)
	if altHead == "" {
		altHead = gh.guessCommitByZIP(plugin.name, base)
	}

	u := *plugin.url
	u.Path = name

	commits := []struct {
		parts   []string
		feed    *GithubFeed
		section string
	}{
		{[]string{"commits", "master"}, nil, "\n - master commits:\n"},
		{[]string{"tags"}, nil, "\n - tags:\n"},
	}
	if head != "master" {
		commit := commits[0]
		commit.parts = []string{"commits", head}
		commit.section = "\n - commits:\n"
		commits = append(commits, commit)
	}

	var wg sync.WaitGroup
	wg.Add(len(commits))
	for i := range commits {
		go func(j int) {
			commits[j].feed = gh.getCommits(&u, commits[j].parts...)
			wg.Done()
		}(i)
	}
	wg.Wait()

	buf := bytes.NewBuffer(nil)

	fmt.Fprintf(buf, "\n\n## %s - %s\n", plugin.name, plugin.url)

	for i := range commits {
		c := &commits[i]
		if c.feed == nil || len(c.feed.Entry) == 0 {
			continue
		}

		fmt.Fprintln(buf, c.section)
		for _, entry := range c.feed.Entry {
			mark := " "
			if entry.ID == head || entry.ID == altHead {
				mark = "*"
			}
			fmt.Fprintf(buf, "  %s%.10s %s %s\n", mark, entry.ID, entry.Updated, entry.Title)
		}
	}

	return buf.String()
}
