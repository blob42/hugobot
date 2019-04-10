package export

import (
	"git.sp4ke.com/sp4ke/hugobot/v3/feeds"
	"git.sp4ke.com/sp4ke/hugobot/v3/posts"
	"path"

	"github.com/gobuffalo/flect"
)

func NewsletterPostLayout(exp Map, feed feeds.Feed, post posts.Post) error {
	section := path.Base(flect.Singularize(feed.Section))
	if feed.Section == "bulletin/newsletters" {
		exp["layout"] = section
	}

	return nil
}
