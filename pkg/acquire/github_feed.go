package acquire

import (
	"encoding/xml"
	"log"
	"net/http"
	"net/url"
	"strings"
	"sync"

	vhttp "github.com/mgumz/vopher/pkg/http"
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
	resp, done, err := vhttp.Request(http.MethodGet, feedURL.String())
	if err != nil {
		log.Printf("error: %q %v", feedURL.String(), err)
		return nil
	}
	defer done()

	// 404 is expected: either "master" or "main" does not exist
	if resp.StatusCode == http.StatusNotFound {
		return nil
	}
	if resp.StatusCode != http.StatusOK {
		msg := resp.Status
		if resp.StatusCode == http.StatusForbidden || resp.StatusCode == http.StatusTooManyRequests {
			msg += " (rate limited?"
			if ra := resp.Header.Get("Retry-After"); ra != "" {
				msg += " retry after " + ra + "s"
			}
			msg += ")"
		}
		log.Printf("error: %q %s", feedURL.String(), msg)
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
		title := strings.TrimSpace(entry.Title)
		feed.Entry[i].ID = feedEntryID(entry.ID)
		feed.Entry[i].Title = title
	}

	return feed
}

// feedEntryID returns the commit-id or the tag name of a feed entry-id:
//
//	tag:github.com,2008:Grit::Commit/<commit>
//	tag:github.com,2008:Repository/<repo-id>/<tag>
//
// The tag name may contain "/", eg. "firestore/v1.27.0".
func feedEntryID(id string) string {
	if _, rest, ok := strings.Cut(id, ":Repository/"); ok {
		if _, tag, ok := strings.Cut(rest, "/"); ok {
			return tag
		}
	}
	return id[strings.LastIndex(id, "/")+1:]
}
