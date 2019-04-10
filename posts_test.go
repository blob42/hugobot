package main

import (
	"git.sp4ke.com/sp4ke/hugobot/v3/posts"
	"log"
	"os"
	"testing"
)

func TestGetPosts(t *testing.T) {
	posts, err := posts.ListPosts()
	if err != nil {
		t.Error(err)
	}
	log.Println(posts)
	for _, p := range posts {
		t.Logf("%s <---- %s", p.Title, p.Feed.Name)
	}

}

func TestMain(m *testing.M) {
	code := m.Run()

	defer DB.Handle.Close()
	os.Exit(code)

}
