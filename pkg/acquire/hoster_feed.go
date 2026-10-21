package acquire

import (
	"encoding/xml"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"
	"sync"

	vhttp "github.com/mgumz/vopher/pkg/http"
)

// atomFeed is a minimal atom-parser sufficient to extract only what we need
// from the commit- and tag-feeds of a git hoster
type atomFeed struct {
	XMLName xml.Name `xml:"feed"`
	Entry   []struct {
		Title   string `xml:"title"`
		Updated string `xml:"updated"`
		ID      string `xml:"id"`
	} `xml:"entry"`
}

type feedSection struct {
	url     url.URL
	entryID func(id string) string // commit-id or tag name of an entry-id
	feed    *atomFeed
	section string
}
type feedSections []feedSection

func buildFeeds(hoster Hoster, repo *url.URL, branch, head string) feedSections {

	commits := feedSections{}

	if branch != "" {
		commits = append(commits,
			feedSection{hoster.CommitsFeedURL(repo, branch), commitID, nil, "\n - " + branch + " commits:\n"})
	} else {
		commits = append(commits,
			feedSection{hoster.CommitsFeedURL(repo, "master"), commitID, nil, "\n - master commits:\n"},
			feedSection{hoster.CommitsFeedURL(repo, "main"), commitID, nil, "\n - main commits:\n"},
		)
		if head != "master" && head != "main" {
			commits = append(commits,
				feedSection{hoster.CommitsFeedURL(repo, head), commitID, nil, "\n - commits:\n"})
		}
	}
	commits = append(commits, feedSection{hoster.TagsFeedURL(repo), hoster.TagName, nil, "\n - tags:\n"})

	return commits
}

func (feeds feedSections) fetch() {

	wg := sync.WaitGroup{}
	wg.Add(len(feeds))

	for i := range feeds {
		go func(j int) {
			feeds[j].feed = fetchFeed(feeds[j].url, feeds[j].entryID)
			wg.Done()
		}(i)
	}

	wg.Wait()
}

// fetchFeed fetches and parses the atom feed at feedURL. The entry-ids are
// reduced by entryID.
func fetchFeed(feedURL url.URL, entryID func(id string) string) *atomFeed {

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

	feed, err := parseFeed(resp.Body, entryID)
	if err != nil {
		log.Printf("error: %q %v", feedURL.String(), err)
		return nil
	}
	if len(feed.Entry) == 0 {
		return nil
	}
	return feed
}

// parseFeed parses the atom feed in r and reduces the entry-ids by entryID
func parseFeed(r io.Reader, entryID func(id string) string) (*atomFeed, error) {

	feed := new(atomFeed)
	if err := xml.NewDecoder(r).Decode(feed); err != nil {
		return nil, err
	}

	for i, entry := range feed.Entry {
		title := strings.TrimSpace(entry.Title)
		feed.Entry[i].ID = entryID(entry.ID)
		feed.Entry[i].Title = title
	}

	return feed, nil
}

// commitID returns the part of the entry-id after the last "/", which is the
// commit-id in the commits feeds of GitHub, Forgejo and GitLab
func commitID(id string) string {
	return id[strings.LastIndex(id, "/")+1:]
}
