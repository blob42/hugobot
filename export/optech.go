package export

import (
	"hugobot/feeds"
	"hugobot/posts"
	"fmt"
	"net/url"

	_ "github.com/fatih/structs"
)

// This happens on exported posts
func OptechExport(exp Map, feed feeds.Feed, post posts.Post) error {
	if feed.Name == "optech" {
		// Export link to newsletter

		base, err := url.Parse(feed.Url)
		if err != nil {
			return err
		}
		base, err = url.Parse(fmt.Sprintf("%s://%s", base.Scheme, base.Host))
		if err != nil {
			return err
		}

		postLink, err := url.Parse(post.Link)
		if err != nil {
			return err
		}

		link := base.ResolveReference(postLink)

		exp["link"] = link.String()

		// Export GUID
		exp["guid"] = post.JsonData["GUID"]
	}
	return nil
}
