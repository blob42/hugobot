package export

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"git.blob42.xyz/blob42/hugobot/v3/config"
	"git.blob42.xyz/blob42/hugobot/v3/encoder"
	"git.blob42.xyz/blob42/hugobot/v3/feeds"
	"git.blob42.xyz/blob42/hugobot/v3/filters"
	"git.blob42.xyz/blob42/hugobot/v3/posts"
	"git.blob42.xyz/blob42/hugobot/v3/types"
	"git.blob42.xyz/blob42/hugobot/v3/utils"
)

var PostMappers []PostMapper
var FeedMappers []FeedMapper

type Map map[string]interface{}

type PostMapper func(Map, feeds.Feed, posts.Post) error
type FeedMapper func(Map, feeds.Feed) error

// Exported version of a post
type PostExport struct {
	ID        int64     `json:"id"`
	Title     string    `json:"title"`
	Link      string    `json:"link"`
	Published time.Time `json:"published"`
	Content   string    `json:"content"`
}

type PostMap map[int64]Map

type FeedExport struct {
	Name       string           `json:"name"`
	Section    string           `json:"section"`
	Categories types.StringList `json:"categories"`
	Posts      PostMap          `json:"posts"`
}

type HugoExporter struct{}

func (he HugoExporter) Handle(feed feeds.Feed) error {
	return he.export(feed)
}

func (he HugoExporter) export(feed feeds.Feed) error {
	log.Printf("Exporting %s to %s", feed.Name, config.HugoData())

	posts, err := posts.GetPostsByFeedId(feed.FeedID)
	if err != nil {
		return err
	}

	if len(posts) == 0 {
		log.Printf("nothing to export")
		return nil
	}

	// Run filters on posts
	for _, p := range posts {
		filters.RunPostFilterHooks(feed, p)
	}

	// Dir and filename
	dirPath := filepath.Join(config.HugoData(), feed.Section)
	cleanFeedName := strings.Replace(feed.Name, "/", "-", -1)
	filePath := filepath.Join(dirPath, cleanFeedName+".json")

	err = utils.Mkdir(dirPath)
	if err != nil {
		return err
	}

	feedExp := Map{
		"name":       feed.Name,
		"section":    feed.Section,
		"categories": feed.Categories,
	}

	runFeedMappers(feedExp, feed)

	postsMap := make(PostMap)
	for _, p := range posts {
		exp := Map{
			"id":        p.PostID,
			"title":     p.Title,
			"link":      p.Link,
			"published": p.Published,
			"updated":   p.Updated,
			//"content":   p.Content,
		}
		runPostMappers(exp, feed, *p)

		postsMap[p.PostID] = exp
	}
	feedExp["posts"] = postsMap

	outputFile, err := os.Create(filePath)
	defer outputFile.Close()
	if err != nil {
		return err
	}

	exportEncoder := encoder.NewExportEncoder(outputFile, encoder.JSON)
	exportEncoder.Encode(feedExp)
	//jsonEnc.Encode(feedExp)

	// Handle feeds which export posts individually as hugo posts
	// Like bulletin

	if feed.ExportPosts {
		for _, p := range posts {

			exp := map[string]interface{}{
				"id":           p.PostID,
				"title":        p.Title,
				"name":         feed.Name,
				"author":       p.Author,
				"description":  p.PostDescription,
				"externalLink": feed.UseExternalLink,
				"display_name": feed.DisplayName,
				"publishdate":  p.Published,
				"date":         p.Updated,
				"issuedate":    utils.NextThursday(p.Updated),
				"use_data":     true,
				"slug":         p.ShortID,
				"link":         p.Link,
				// Content is written in the post
				"content":    p.Content,
				"categories": feed.Categories,
				"tags":       strings.Split(p.Tags, ","),
			}

			if feed.Publications != "" {
				exp["publications"] = strings.Split(feed.Publications, ",")
			}

			runPostMappers(exp, feed, *p)

			dirPath := filepath.Join(config.HugoContent(), feed.Section)
			cleanFeedName := strings.Replace(feed.Name, "/", "-", -1)
			fileName := fmt.Sprintf("%s-%s.md", cleanFeedName, p.ShortID)
			filePath := filepath.Join(dirPath, fileName)

			outputFile, err := os.Create(filePath)

			defer outputFile.Close()

			if err != nil {
				return err
			}

			exportEncoder := encoder.NewExportEncoder(outputFile, encoder.TOML)
			exportEncoder.Encode(exp)
		}
	}
	return nil
}

// Runs in goroutine
func (he HugoExporter) Export(feed feeds.Feed) {
	err := he.export(feed)
	if err != nil {
		log.Fatal(err)
	}
}

func NewHugoExporter() HugoExporter {
	// Make sure path exists
	err := utils.Mkdir(config.HugoData())
	if err != nil {
		log.Fatal(err)
	}
	return HugoExporter{}
}

func runPostMappers(e Map, f feeds.Feed, p posts.Post) {
	for _, fn := range PostMappers {
		err := fn(e, f, p)
		if err != nil {
			log.Print(err)
		}
	}
}

func runFeedMappers(e Map, f feeds.Feed) {
	for _, fn := range FeedMappers {
		err := fn(e, f)
		if err != nil {
			log.Print(err)
		}
	}
}

func RegisterPostMapper(mapper PostMapper) {
	PostMappers = append(PostMappers, mapper)
}

func RegisterFeedMapper(mapper FeedMapper) {
	FeedMappers = append(FeedMappers, mapper)
}

func init() {
	RegisterPostMapper(BulletinExport)
	RegisterPostMapper(NewsletterPostLayout)
	RegisterPostMapper(RFCExport)
	RegisterPostMapper(ReleaseExport)
}
