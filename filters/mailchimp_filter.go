package filters

import (
	"git.blob42.xyz/blob42/hugobot/v3/feeds"
	"git.blob42.xyz/blob42/hugobot/v3/posts"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

const (
	PreviewTextSel = ".mcnPreviewText"
)

var (
	RemoveSelectors = []string{"style", ".footerContainer", "#awesomewrap", "#templatePreheader", "img", "head"}
)

func mailChimpFilter(feed feeds.Feed, post *posts.Post) error {

	// Nothing to do for empty content
	if post.PostDescription == post.Content &&
		post.Content == "" {
		return nil
	}

	// Same content in both
	if post.PostDescription == post.Content {
		post.PostDescription = ""

	}

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(post.Content))
	if err != nil {
		return err
	}

	sel := doc.Find(strings.Join(RemoveSelectors, ","))
	sel.Remove()

	post.Content, err = doc.Html()

	return err
}

func extractPreviewText(feed feeds.Feed, post *posts.Post) error {
	// Ignore filled description
	if post.PostDescription != "" {
		return nil
	}

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(post.Content))
	if err != nil {
		return err
	}

	sel := doc.Find(PreviewTextSel)
	post.PostDescription = sel.Text()
	return nil
}

func init() {
	RegisterPostFilterHook(mailChimpFilter)
	RegisterPostFilterHook(extractPreviewText)
}
