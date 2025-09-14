package acquire

import (
	"encoding/xml"
	"log"
	"net/http"
	"net/url"
	"strings"
	"sync"
)

// GithubFeed is a minimal atom-parser sufficient to extract only what we need
// from Github
type githubFeed struct {
	XMLName xml.Name `xml:"feed"`
	Entry   []struct {
		Title   string `xml:"title"`
		Updated string `xml:"updated"`
		ID      string `xml:"id"`
	} `xml:"entry"`
}

type ghFeed struct {
	parts   []string
	feed    *githubFeed
	section string
}
type ghFeeds []ghFeed

func buildFeeds(branch, head string) ghFeeds {

	commits := ghFeeds{}

	if branch != "" {
		commits = append(commits,
			ghFeed{[]string{"commits", branch}, nil, "\n - " + branch + " commits:\n"})
	} else {
		commits = append(commits,
			ghFeed{[]string{"commits", "master"}, nil, "\n - master commits:\n"},
			ghFeed{[]string{"commits", "main"}, nil, "\n - main commits:\n"},
		)
		if head != "master" && head != "main" {
			commits = append(commits,
				ghFeed{[]string{"commits", head}, nil, "\n - commits:\n"})
		}
	}
	commits = append(commits, ghFeed{[]string{"tags"}, nil, "\n - tags:\n"})

	return commits
}

func (feeds ghFeeds) fetch(gh Github, repo *url.URL) {

	wg := sync.WaitGroup{}
	wg.Add(len(feeds))

	for i := range feeds {
		go func(j int) {
			feeds[j].feed = gh.GetCommits(repo, feeds[j].parts...)
			wg.Done()
		}(i)
	}

	wg.Wait()
}

// GetCommits returns the GithubFeed of the repository
func (gh Github) GetCommits(repo *url.URL, parts ...string) *githubFeed {

	feedURL := gh.FeedURL(repo, parts...)
	resp, err := http.Get(feedURL.String())
	if err != nil {
		log.Printf("error: %q %v", feedURL.String(), err)
		return nil
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil
	}
	if resp.StatusCode != http.StatusOK {
		log.Printf("error: %q %v", feedURL.String(), err)
		return nil
	}

	xmldec := xml.NewDecoder(resp.Body)
	feed := new(githubFeed)

	if err = xmldec.Decode(feed); err != nil {
		log.Printf("error: %q %v", feedURL.String(), err)
		return nil
	}

	if len(feed.Entry) == 0 {
		return nil
	}

	for i, entry := range feed.Entry {
		idx := strings.LastIndex(entry.ID, "/")
		id := entry.ID[idx+1:]
		title := strings.TrimSpace(entry.Title)
		feed.Entry[i].ID = id
		feed.Entry[i].Title = title
	}

	return feed
}
