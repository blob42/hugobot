package github

import (
	"git.blob42.xyz/blob42/hugobot/v3/config"
	"context"

	"github.com/google/go-github/github"
	"golang.org/x/oauth2"
)

func Auth(ctx context.Context) *github.Client {

	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: config.C.GithubAccessToken},
	)

	tc := oauth2.NewClient(ctx, ts)
	client := github.NewClient(tc)

	return client
}
