package export

import (
	"git.blob42.xyz/blob42/hugobot/v3/feeds"
	"git.blob42.xyz/blob42/hugobot/v3/posts"
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
