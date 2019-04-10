package posts

import (
	"git.sp4ke.com/sp4ke/hugobot/v3/db"
	"git.sp4ke.com/sp4ke/hugobot/v3/feeds"
	"git.sp4ke.com/sp4ke/hugobot/v3/types"
	"git.sp4ke.com/sp4ke/hugobot/v3/utils"
	"errors"
	"fmt"
	"log"
	"strconv"
	"time"

	sqlite3 "github.com/mattn/go-sqlite3"
)

const (
	DBPostsSchema = `CREATE TABLE IF NOT EXISTS posts (
		post_id INTEGER PRIMARY KEY,
		feed_id INTEGER NOT NULL,
		title TEXT DEFAULT '',
		description TEXT DEFAULT '',
		link TEXT NOT NULL,
		updated timestamp NOT NULL,
		published timestamp NOT NULL,
		author TEXT DEFAULT '',
		content TEXT DEFAULT '',
		tags TEXT DEFAULT '',
		json_data BLOB DEFAULT '',
		short_id TEXT UNIQUE,
		FOREIGN KEY (feed_id) REFERENCES feeds(feed_id)
	)`
)

var (
	ErrDoesNotExist  = errors.New("does not exist")
	ErrAlreadyExists = errors.New("already exists")
)

var DB = db.DB

type Post struct {
	PostID          int64         `josn:"id" db:"post_id"`
	Title           string        `json:"title"`
	PostDescription string        `json:"description" db:"post_description"`
	Link            string        `json:"link"`
	Updated         time.Time     `json:"updated"`
	Published       time.Time     `json:"published"`
	Author          string        `json:"author"`
	Content         string        `json:"content"`
	Tags            string        `json:"tags"`
	ShortID         string        `json:"short_id" db:"short_id"`
	JsonData        types.JsonMap `json:"data" db:"json_data"`

	feeds.Feed
}

// Writes with provided short id
func (post *Post) WriteWithShortId(feedId int64, shortId interface{}) error {
	var shortid string

	switch v := shortId.(type) {
	case int:
		shortid = strconv.Itoa(v)
	case int64:
		shortid = strconv.Itoa(int(v))
	case string:
		shortid = v
	default:
		return fmt.Errorf("Cannot convert %v to string", shortId)

	}
	return write(post, feedId, shortid)
}

// Auto generates shortId
func (post *Post) Write(feedId int64) error {

	shortId, err := utils.GetSIDGenerator().Generate()
	if err != nil {
		return err
	}

	return write(post, feedId, shortId)
}

func write(post *Post, feedId int64, shortId string) error {
	const query = `INSERT OR REPLACE INTO posts (
feed_id,
title,
description,
link,
updated,
published,
author,
content,
json_data,
short_id,
tags
)

VALUES(
?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?
)
`

	_, err := DB.Handle.Exec(query,
		feedId,
		post.Title,
		post.PostDescription,
		post.Link,
		post.Updated,
		post.Published,
		post.Author,
		post.Content,
		post.JsonData,
		shortId,
		post.Tags,
	)

	sqlErr, isSqlErr := err.(sqlite3.Error)
	if isSqlErr && sqlErr.Code == sqlite3.ErrConstraint {
		return fmt.Errorf("%+v --- %s ", sqlErr, post.Title)
	}

	if err != nil {
		return err
	}

	return nil
}

func ListPosts() ([]Post, error) {
	const query = `SELECT * FROM posts JOIN feeds ON posts.feed_id = feeds.feed_id`
	var posts []Post
	err := DB.Handle.Select(&posts, query)
	if err != nil {
		return nil, err
	}
	return posts, nil

}

func GetPostsByFeedId(feedId int64) ([]*Post, error) {

	const query = `SELECT 
post_id,
feed_id,
title,
description AS post_description,
link,
updated,
published,
author,
content,
tags,
json_data,
short_id
		FROM posts WHERE feed_id = ?`

	var posts []*Post

	err := DB.Handle.Select(&posts, query, feedId)
	if err != nil {
		return nil, err
	}

	return posts, nil
}

func init() {
	_, err := DB.Handle.Exec(DBPostsSchema)
	if err != nil {
		log.Fatal(err)
	}
}
