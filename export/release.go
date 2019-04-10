package export

import (
	"hugobot/feeds"
	"hugobot/posts"
)

//
func ReleaseExport(exp Map, feed feeds.Feed, post posts.Post) error {
	if feed.Section == "bulletin/releases" {
		exp["data"] = post.JsonData
	}
	return nil
}
