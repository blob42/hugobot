package filters

import (
	"git.sp4ke.com/sp4ke/hugobot/v3/feeds"
	"git.sp4ke.com/sp4ke/hugobot/v3/posts"
	"log"
)

type FilterHook func(feed feeds.Feed, post *posts.Post) error

var (
	PostFilters []FilterHook
)

func RegisterPostFilterHook(hook FilterHook) {
	PostFilters = append(PostFilters, hook)
}

func RunPostFilterHooks(feed feeds.Feed, post *posts.Post) {
	for _, h := range PostFilters {
		err := h(feed, post)
		if err != nil {
			log.Fatal(err)
		}
	}
}
