package handlers

import (
	"hugobot/feeds"
	"hugobot/github"
	"hugobot/posts"
	"hugobot/utils"
	"context"
	"errors"
	"log"
	"time"

	githubApi "github.com/google/go-github/github"
)

const (
	NoReleaseForProjectTreshold = 10
)

type GHRelease struct {
	ProjectID int64     `json:"project_id"`
	ReleaseID int64     `json:"release_id"`
	TagID     string    `json:"tag_id"`
	Name      string    `json:"name"`
	IsTagOnly bool      `json:"is_tag_only"`
	Date      time.Time `json:"commit_date"`
	TarBall   string    `json:"tar_ball"`
	Link      string    `json:"link"`
	Owner     string    `json:"owner"`
	Repo      string    `json:"repo"`
}

type GHReleaseHandler struct {
	ctx      context.Context
	ghClient *githubApi.Client
}

func (handler GHReleaseHandler) Handle(feed feeds.Feed) error {
	posts, err := handler.FetchSince(feed.Url, feed.LastRefresh)
	if err != nil {
		return err
	}

	if posts == nil {
		log.Printf("No new posts in feed <%s>", feed.Name)
	}

	for _, p := range posts {
		var err error
		isTagOnly, ok := p.JsonData["is_tag_only"].(bool)

		if !ok {
			return errors.New("could not convert is_tag_only to bool")
		}

		if isTagOnly {
			err = p.Write(feed.FeedID)
		} else {
			err = p.WriteWithShortId(feed.FeedID, p.JsonData["release_id"])
		}
		if err != nil {
			return err
		}
	}

	return nil
}

func (handler GHReleaseHandler) FetchSince(url string,
	after time.Time) ([]*posts.Post, error) {
	var results []*posts.Post

	log.Printf("Fetching GH release %s since %v", url, after)

	owner, repo := github.ParseOwnerRepo(url)

	project, resp, err := handler.ghClient.Repositories.Get(handler.ctx, owner, repo)
	if resp == nil {
		return nil, errors.New("No response")
	}

	github.RespMiddleware(resp)
	if err != nil {
		return nil, err
	}

	// Handle releases first, if project has no releases use tags instead
	listReleasesOptions := &githubApi.ListOptions{
		PerPage: 100,
	}

	releases, resp, err := handler.ghClient.Repositories.ListReleases(
		handler.ctx, owner, repo, listReleasesOptions,
	)
	github.RespMiddleware(resp)
	if err != nil {
		return nil, err
	}

	// If no releases use tags
	if len(releases) <= 0 {
		log.Println("no releases, using tags")
		var allTags []*githubApi.RepositoryTag
		// Handle tags first
		listTagOptions := &githubApi.ListOptions{PerPage: 100}

		for {
			tags, resp, err := handler.ghClient.Repositories.ListTags(
				handler.ctx, owner, repo, listTagOptions,
			)

			if err != nil {
				return nil, err
			}

			allTags = append(allTags, tags...)
			if resp.NextPage == 0 {
				break
			}

			listTagOptions.Page = resp.NextPage
		}

		for _, tag := range allTags {
			//var release *githubApi.RepositoryRelease
			//
			commit, resp, err := handler.ghClient.Repositories.GetCommit(
				handler.ctx, owner, repo, tag.GetCommit().GetSHA(),
			)
			github.RespMiddleware(resp)
			if err != nil {
				return nil, err
			}

			if commit.GetCommit().GetCommitter().GetDate().Before(after) {
				break
			}

			ghRelease := GHRelease{
				ProjectID: project.GetID(),
				TagID:     tag.GetName(),
				IsTagOnly: true,
				Date:      commit.GetCommit().GetCommitter().GetDate(),
				TarBall:   tag.GetTarballURL(),
				Owner:     owner,
				Repo:      repo,
			}

			post := &posts.Post{}
			post.Title = tag.GetName()
			post.Link = ghRelease.TarBall
			post.Published = ghRelease.Date
			post.Updated = post.Published
			post.JsonData = utils.StructToJsonMap(ghRelease)
			post.Author = commit.GetAuthor().GetName()

			results = append(results, post)
		}
	} else {

		for _, release := range releases {

			ghRelease := GHRelease{
				ProjectID: project.GetID(),
				ReleaseID: release.GetID(),
				Name:      release.GetName(),
				TagID:     release.GetTagName(),
				IsTagOnly: false,
				Date:      release.GetCreatedAt().Time,
				TarBall:   release.GetTarballURL(),
				Owner:     owner,
				Repo:      repo,
			}

			post := &posts.Post{}
			post.Title = release.GetName()
			post.Link = release.GetHTMLURL()
			post.Published = release.GetPublishedAt().Time
			post.Updated = release.GetPublishedAt().Time
			post.JsonData = utils.StructToJsonMap(ghRelease)
			post.Author = release.GetAuthor().GetName()
			post.Content = release.GetBody()

			results = append(results, post)
		}

	}

	return results, nil
}

func NewGHReleaseHandler() FormatHandler {
	ctxb := context.Background()
	client := github.Auth(ctxb)

	return GHReleaseHandler{
		ctx:      ctxb,
		ghClient: client,
	}
}
