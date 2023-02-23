package main

import (
	"git.blob42.xyz/blob42/hugobot/v3/handlers"
	"fmt"
	"os"
	"testing"
	"time"
)

const (
	rssTestFeed  = "https://bitcointechweekly.com/index.xml"
	rssTestFeed2 = "https://bitcoinops.org/feed.xml"
)

func TestFetch(t *testing.T) {
	handler := handlers.NewRSSHandler()
	when, _ := time.Parse("Jan 2006", "Jun 2018")
	res, err := handler.FetchSince(rssTestFeed2, when)
	if err != nil {
		t.Error(err)
	}

	for i, post := range res {
		f, err := os.Create(fmt.Sprintf("%d.html", i))
		if err != nil {
			t.Error(err)
		}
		defer f.Close()
		_, err = f.WriteString(post.Content)
		if err != nil {
			t.Error(err)
		}
	}

}
