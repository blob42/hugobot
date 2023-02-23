package export

import (
	"git.blob42.xyz/blob42/hugobot/v3/feeds"
	"git.blob42.xyz/blob42/hugobot/v3/posts"
)

//
func ReleaseExport(exp Map, feed feeds.Feed, post posts.Post) error {
	if feed.Section == "bulletin/releases" {
		exp["data"] = post.JsonData
	}
	return nil
}
