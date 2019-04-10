package main

import (
	"git.sp4ke.com/sp4ke/hugobot/v3/feeds"
	"git.sp4ke.com/sp4ke/hugobot/v3/handlers"
	"git.sp4ke.com/sp4ke/hugobot/v3/posts"
	"fmt"
	"log"
	"time"

	cli "gopkg.in/urfave/cli.v1"
)

var fetchCmd = cli.Command{
	Name:    "fetch",
	Aliases: []string{"f"},
	Usage:   "Fetch data from feed",
	Flags: []cli.Flag{
		cli.StringFlag{
			Name:  "since",
			Usage: "Fetch data since `TIME`, defaults to last refresh time",
		},
	},
	Action: fetchFeeds,
}

var feedsCmdGroup = cli.Command{
	Name:  "feeds",
	Usage: "Feeds related commands. default: list feeds",
	Flags: []cli.Flag{
		cli.IntFlag{
			Name:  "id,i",
			Value: 0,
			Usage: "Feeds `id`",
		},
	},
	Subcommands: []cli.Command{
		fetchCmd,
	},
	Action: listFeeds,
}

func fetchFeeds(c *cli.Context) {
	var result []*posts.Post

	fList, err := getFeeds(c.Parent())
	if err != nil {
		log.Fatal(err)
	}

	for _, f := range fList {
		var handler handlers.FormatHandler
		handler = handlers.GetFormatHandler(*f)

		if c.IsSet("since") {
			// Parse time
			t, err := time.Parse(time.UnixDate, c.String("since"))
			if err != nil {
				log.Fatal(err)
			}
			result, err = handler.FetchSince(f.Url, t)

		} else {
			result, err = handler.FetchSince(f.Url, f.LastRefresh)
		}

		if err != nil {
			log.Fatal(err)
		}

		for _, post := range result {
			log.Printf("%s (updated: %s)", post.Title, post.Updated)
		}
		log.Println("Total: ", len(result))

	}

}

func listFeeds(c *cli.Context) {
	fList, err := getFeeds(c)
	if err != nil {
		log.Fatal(err)
	}

	for _, f := range fList {
		fmt.Println(f)
	}
}

func getFeeds(c *cli.Context) ([]*feeds.Feed, error) {
	var fList []*feeds.Feed
	var err error

	if c.IsSet("id") {
		feed, err := feeds.GetById(c.Int64("id"))
		if err != nil {
			return nil, err
		}

		fList = append(fList, feed)
	} else {
		fList, err = feeds.ListFeeds()
		if err != nil {
			return nil, err
		}

	}

	return fList, nil

}
