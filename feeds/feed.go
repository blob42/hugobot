package feeds

import (
	"git.blob42.xyz/blob42/hugobot/v3/db"
	"git.blob42.xyz/blob42/hugobot/v3/types"
	"errors"
	"log"
	"time"

	sqlite3 "github.com/mattn/go-sqlite3"
)

//sqlite> SELECT feeds.name, url, feed_formats.name AS format_name from feeds JOIN feed_formats ON feeds.format = feed_formats.id;
//
var DB = db.DB

const (
	DBFeedSchema = `CREATE TABLE IF NOT EXISTS feeds (
		feed_id INTEGER PRIMARY KEY,
		name TEXT NOT NULL UNIQUE,
		display_name TEXT DEFAULT '',
		publications TEXT DEFAULT '',
		section TEXT DEFAULT '',
		categories TEXT DEFAULT '',
		description TEXT DEFAULT '',
		url TEXT NOT NULL,
		export_posts INTEGER DEFAULT 0,
		last_refresh timestamp DEFAULT -1,
		created timestamp DEFAULT (strftime('%s')),
		interval INTEGER DEFAULT 60,
		format INTEGER NOT NULL DEFAULT 0,
		serial_run INTEGER DEFAULT 0,
		use_external_link INTEGER DEFAULT 0,
		FOREIGN KEY (format) REFERENCES feed_formats(id)


	)`

	DBFeedFormatsSchema = `CREATE TABLE IF NOT EXISTS feed_formats (
		id INTEGER PRIMARY KEY,
		format_name TEXT NOT NULL UNIQUE
	)`
)

const (
	QDeleteFeedById = `DELETE FROM feeds WHERE feed_id = ?`
	QGetFeed        = `SELECT * FROM feeds WHERE feed_id = ?`
	QGetFeedByName  = `SELECT * FROM feeds WHERE name = ?`
	QGetFeedByURL   = `SELECT * FROM feeds WHERE url = ?`
	QListFeeds      = `SELECT 
						feeds.feed_id,
						feeds.name,
						feeds.display_name,
						feeds.publications,
						feeds.section,
						feeds.categories,
						feeds.description,
						feeds.url,
						feeds.last_refresh,
						feeds.created,
						feeds.format,
						feeds.serial_run,
						feeds.use_external_link,
						feeds.interval,
						feeds.export_posts,
						feed_formats.format_name
						 FROM feeds
						JOIN feed_formats ON feeds.format = feed_formats.id`
)

var (
	ErrDoesNotExist  = errors.New("does not exist")
	ErrAlreadyExists = errors.New("already exists")
)

type FeedFormat int

// Feed Formats
const (
	FormatRSS FeedFormat = iota
	FormatHTML
	FormatJSON
	FormatTweet
	FormatRFC
	FormatGHRelease
)

var FeedFormats = map[FeedFormat]string{
	FormatRSS:       "RSS",
	FormatHTML:      "HTML",
	FormatJSON:      "JSON",
	FormatTweet:     "TWEET",
	FormatRFC:       "RFC",
	FormatGHRelease: "GithubRelease",
}

type Feed struct {
	FeedID       int64            `json:"id" db:"feed_id"`
	Name         string           `json:"name" db:"name"`
	Section      string           `json:"section,omitempty"`
	Categories   types.StringList `json:"categories,omitempty"`
	Description  string           `json:"description"`
	Url          string           `json:"url"`
	Format       FeedFormat       `json:"-"`
	FormatString string           `json:"format" db:"format_name"`
	LastRefresh  time.Time        `db:"last_refresh" json:"last_refresh"` // timestamp time.Unix()
	Created      time.Time        `json:"created"`
	DisplayName  string           `db:"display_name"`
	Publications string           `json:"-"`

	// This feed's posts should also be exported individually
	ExportPosts bool `json:"export_posts" db:"export_posts"`

	// Time in seconds between each polling job on the news feed
	Interval float64 `json:"refresh_interval"`

	Serial bool `json:"serial" db:"serial_run"` // Jobs for this feed should run in series

	// Items which only contain summaries and redirect to external content
	// like publications and newsletters
	UseExternalLink bool `json:"use_external_link" db:"use_external_link"`
}

func (f *Feed) Write() error {

	query := `INSERT INTO feeds
		(name, section, categories, url, format)
VALUES(:name, :section, :categories, :url, :format)`

	_, err := DB.Handle.NamedExec(query, f)
	sqlErr, isSqlErr := err.(sqlite3.Error)
	if isSqlErr && sqlErr.Code == sqlite3.ErrConstraint {
		return ErrAlreadyExists
	}

	if err != nil {
		return err
	}

	return nil
}

func (f *Feed) UpdateRefreshTime(time time.Time) error {
	f.LastRefresh = time

	query := `UPDATE feeds SET last_refresh = ? WHERE feed_id = ?`
	_, err := DB.Handle.Exec(query, f.LastRefresh, f.FeedID)
	if err != nil {
		return err
	}

	return nil
}

func GetById(id int64) (*Feed, error) {

	var feed Feed
	err := DB.Handle.Get(&feed, QGetFeed, id)
	if err != nil {
		return nil, err
	}

	feed.FormatString = FeedFormats[feed.Format]

	return &feed, nil
}

func GetByName(name string) (*Feed, error) {

	var feed Feed
	err := DB.Handle.Get(&feed, QGetFeedByName, name)
	if err != nil {
		return nil, err
	}

	feed.FormatString = FeedFormats[feed.Format]

	return &feed, nil
}

func GetByURL(url string) (*Feed, error) {

	var feed Feed
	err := DB.Handle.Get(&feed, QGetFeedByURL, url)
	if err != nil {
		return nil, err
	}

	feed.FormatString = FeedFormats[feed.Format]

	return &feed, nil

}

func ListFeeds() ([]*Feed, error) {
	var feeds []*Feed
	err := DB.Handle.Select(&feeds, QListFeeds)
	if err != nil {
		return nil, err
	}

	return feeds, nil
}

func DeleteById(id int) error {

	// If id does not exists return warning
	var feedToDelete Feed
	err := DB.Handle.Get(&feedToDelete, QGetFeed, id)
	if err != nil {
		return ErrDoesNotExist
	}

	_, err = DB.Handle.Exec(QDeleteFeedById, id)
	if err != nil {
		return err
	}

	return nil
}

// Returns true if the feed should be refreshed
func (feed *Feed) ShouldRefresh() (float64, bool) {
	lastRefresh := feed.LastRefresh
	delta := time.Since(lastRefresh).Seconds() // Delta since last refresh
	//log.Printf("%s delta %f >= interval %f ?", feed.Name, delta, feed.Interval)
	//
	//
	//log.Printf("refresh %s in %.0f seconds", feed.Name, feed.Interval-delta)
	return delta, delta >= feed.Interval
}

func init() {
	_, err := DB.Handle.Exec(DBFeedSchema)
	if err != nil {
		log.Fatal(err)
	}

	_, err = DB.Handle.Exec(DBFeedFormatsSchema)
	if err != nil {
		log.Fatal(err)
	}

	// Populate feed formats
	query := `INSERT INTO feed_formats (id, format_name) VALUES (?, ?)`
	for k, v := range FeedFormats {
		_, err := DB.Handle.Exec(query, k, v)
		if err != nil {

			sqlErr, ok := err.(sqlite3.Error)
			if ok && sqlErr.ExtendedCode == sqlite3.ErrConstraintUnique {
				log.Panic(err)
			}

			if !ok {
				log.Panic(err)
			}
		}

	}

}
