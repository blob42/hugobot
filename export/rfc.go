package export

import (
	"git.sp4ke.com/sp4ke/hugobot/v3/feeds"
	"git.sp4ke.com/sp4ke/hugobot/v3/posts"
)

// TODO: This happend in the main export file
func RFCExport(exp Map, feed feeds.Feed, post posts.Post) error {
	if feed.Section == "bulletin/rfc" {
		exp["data"] = post.JsonData

	}
	return nil
}
