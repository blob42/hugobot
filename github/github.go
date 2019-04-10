package github

import (
	"log"
	"net/url"
	"strings"
	"time"

	"github.com/google/go-github/github"
)

func ParseOwnerRepo(Url string) (owner, repo string) {
	url, err := url.Parse(Url)

	if err != nil {
		panic(err)
	}

	parts := strings.Split(strings.TrimPrefix(url.Path, "/"), "/")
	owner = parts[0]
	repo = parts[1]

	return owner, repo
}

func RespMiddleware(resp *github.Response) {
	if resp == nil {
		return
	}

	log.Printf("Rate remaining: %d/%d (reset: %s)", resp.Rate.Remaining, resp.Rate.Limit, resp.Rate.Reset)
	err := github.CheckResponse(resp.Response)
	if _, ok := err.(*github.RateLimitError); ok {
		log.Printf("HIT RATE LIMIT !!!")
	}
}

type Release struct {
	Version     string    `json:"version"`
	ID          int64     `json:"ID"`
	Date        time.Time `json:"date"`
	Description string    `json:"description"`
	URL         string    `json:"url"`
	ShortURL    string    `json:"short_url"`
	HtmlURL     string    `json:"html_url"`
	Name        string    `json:"name"`
	TagName     string    `json:"tag_name"`
}

type PR struct {
	Title    string    `json:"title"`
	URL      string    `json:"url"`
	HtmlURL  string    `json:"html_url"`
	Number   int       `json:"number"`
	Date     time.Time `json:"date"`
	IssueURL string    `json:"issue_url"`
	Body     string    `json:"body"`
}

type Issue struct {
	Title    string    `json:"title"`
	URL      string    `json:"url"`
	ShortURL string    `json:"short_url"`
	Number   int       `json:"number"`
	State    string    `json:"state"`
	Updated  time.Time `json:"updated"`
	Created  time.Time `json:"created"`
	Comments int       `json:"comments"`
	HtmlURL  string    `json:"html_url"`
	Open     bool      `json:"opened_issue"`
	IsPR     bool      `json:"is_pr"`
	Merged   bool      `json:"merged"` // only for PRs
	MergedAt time.Time `json:"merged_at"`
	IsUpdate bool      `json:"is_update"`
}
