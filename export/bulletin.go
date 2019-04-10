package export

import (
	"hugobot/feeds"
	"hugobot/posts"
	"strings"
)

func BulletinExport(exp Map, feed feeds.Feed, post posts.Post) error {

	bulletinInfo := strings.Split(feed.Section, "/")

	if bulletinInfo[0] == "bulletin" {
		exp["bulletin_type"] = bulletinInfo[1]
	}
	return nil
}
