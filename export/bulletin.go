package export

import (
	"git.blob42.xyz/blob42/hugobot/v3/feeds"
	"git.blob42.xyz/blob42/hugobot/v3/posts"
	"strings"
)

func BulletinExport(exp Map, feed feeds.Feed, post posts.Post) error {

	bulletinInfo := strings.Split(feed.Section, "/")

	if bulletinInfo[0] == "bulletin" {
		exp["bulletin_type"] = bulletinInfo[1]
	}
	return nil
}
