package handlers

import (
	"hugobot/export"
	"hugobot/feeds"
	"hugobot/posts"
	"log"
	"strings"
	"time"

	"github.com/fatih/structs"
	"github.com/mmcdole/gofeed"
)

type RSSHandler struct {
	rssFeed *gofeed.Feed
}

func (handler RSSHandler) Handle(feed feeds.Feed) error {

	posts, err := handler.FetchSince(feed.Url, feed.LastRefresh)
	if err != nil {
		return err
	}

	if posts == nil {
		log.Printf("No new posts in feed <%s>", feed.Name)
	}

	// Write posts to DB
	for _, p := range posts {
		err := p.Write(feed.FeedID)
		if err != nil {
			return err
		}
	}

	return nil
}

func (handler RSSHandler) FetchSince(url string, after time.Time) ([]*posts.Post, error) {
	var err error
	var fetchedPosts []*posts.Post

	log.Printf("Fetching RSS since %v", after)
	fp := gofeed.NewParser()
	handler.rssFeed, err = fp.ParseURL(url)
	if err != nil {
		return nil, err
	}

	for _, item := range handler.rssFeed.Items {
		if item.PublishedParsed.After(after) {
			//log.Println(item.Title)

			post := &posts.Post{}

			if item.Author != nil {
				post.Author = item.Author.Name
			}

			post.Title = item.Title

			// If content is in description
			// store them in reverse in the post
			if len(item.Content) == 0 &&
				len(item.Description) > 0 {
				post.Content = item.Description
				// If content is same as description
			} else if item.Content == item.Description {
				post.Content = item.Content
				post.PostDescription = ""

			} else {
				post.Content = item.Content
				post.PostDescription = item.Description
			}

			post.Link = item.Link

			if item.UpdatedParsed != nil {
				post.Updated = *item.UpdatedParsed
			} else {
				post.Updated = *item.PublishedParsed
			}

			if item.PublishedParsed != nil {
				post.Published = *item.PublishedParsed
			}

			post.Tags = strings.Join(item.Categories, ",")

			item.Content = ""
			item.Description = ""
			post.JsonData = structs.Map(item)

			fetchedPosts = append(fetchedPosts, post)
		}
	}

	return fetchedPosts, nil
}

func NewRSSHandler() FormatHandler {
	return RSSHandler{}
}

func RSSExportMapper(exp export.Map, feed feeds.Feed, post posts.Post) error {
	if feed.Format == feeds.FormatRSS {
		exp["updated"] = post.Updated
	}

	return nil
}

func init() {
	export.RegisterPostMapper(RSSExportMapper)
}
