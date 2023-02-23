package handlers

import (
	"git.blob42.xyz/blob42/hugobot/v3/feeds"
	"git.blob42.xyz/blob42/hugobot/v3/posts"
	"log"
	"time"
)

type JobHandler interface {
	// Main handling function
	Handle(feeds.Feed) error
}

type FormatHandler interface {
	FetchSince(url string, time time.Time) ([]*posts.Post, error)
	JobHandler // Also implements a job handler
}

func GetFormatHandler(feed feeds.Feed) FormatHandler {

	var handler FormatHandler

	switch feed.Format {
	case feeds.FormatRSS:
		handler = NewRSSHandler()
	case feeds.FormatRFC:
		handler = NewRFCHandler()
	case feeds.FormatGHRelease:
		handler = NewGHReleaseHandler()
	default:
		log.Printf("WARNING: No format handler for %s", feed.FormatString)
	}

	return handler
}
