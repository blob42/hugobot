package export

import (
	"git.sp4ke.com/sp4ke/hugobot/v3/feeds"
	"git.sp4ke.com/sp4ke/hugobot/v3/posts"
)

//
func ReleaseExport(exp Map, feed feeds.Feed, post posts.Post) error {
	if feed.Section == "bulletin/releases" {
		exp["data"] = post.JsonData
	}
	return nil
}
