package handlers

import (
	"git.sp4ke.com/sp4ke/hugobot/v3/feeds"
	"git.sp4ke.com/sp4ke/hugobot/v3/github"
	"git.sp4ke.com/sp4ke/hugobot/v3/posts"
	"git.sp4ke.com/sp4ke/hugobot/v3/static"
	"git.sp4ke.com/sp4ke/hugobot/v3/utils"
	"context"
	"fmt"
	"log"
	"regexp"
	"strconv"
	"strings"
	"time"

	githubApi "github.com/google/go-github/github"
)

var (
	ReBIP  = regexp.MustCompile(`bips?[\s-]*(?P<bipId>[0-9]+)`)
	ReBOLT = regexp.MustCompile(`bolt?[\s-]*(?P<boltId>[0-9]+)`)
	ReSLIP = regexp.MustCompile(`slips?[\s-]*(?P<slipId>[0-9]+)`)
)

const (
	BIPLink  = "https://github.com/bitcoin/bips/blob/master/bip-%04d.mediawiki"
	SLIPLink = "https://github.com/satoshilabs/slips/blob/master/slip-%04d.md"
)

const (
	BOLT = 73738971
	BIP  = 14531737
	SLIP = 50844973
)

var RFCTypes = map[int64]string{
	BIP:  "bip",
	BOLT: "bolt",
	SLIP: "slip",
}

type RFCUpdate struct {
	RFCID     int64         `json:"rfcid"`
	RFCType   string        `json:"rfc_type"`   //bip bolt slip
	RFCNumber int           `json:"rfc_number"` // (bip/bolt/slip id)
	Issue     *github.Issue `json:"issue"`
	RFCLink   string        `json:"rfc_link"`
}

type RFCHandler struct {
	ctx      context.Context
	ghClient *githubApi.Client
}

func (handler RFCHandler) Handle(feed feeds.Feed) error {

	posts, err := handler.FetchSince(feed.Url, feed.LastRefresh)
	if err != nil {
		return err
	}

	if posts == nil {
		log.Printf("No new posts in feed <%s>", feed.Name)
	}

	for _, p := range posts {
		// Since RFCs are based on github issues, we use their id as unique
		// id in the local sqlite db
		err := p.WriteWithShortId(feed.FeedID, p.JsonData["rfcid"])
		if err != nil {
			return err
		}

	}

	return nil
}

func (handler RFCHandler) FetchSince(url string, after time.Time) ([]*posts.Post, error) {
	var results []*posts.Post

	log.Printf("Fetching RFC %s since %v", url, after)

	owner, repo := github.ParseOwnerRepo(url)

	project, resp, err := handler.ghClient.Repositories.Get(handler.ctx, owner, repo)
	github.RespMiddleware(resp)

	if err != nil {
		return nil, err
	}

	//All Issues
	var allIssues []*githubApi.Issue
	listIssueOptions := &githubApi.IssueListByRepoOptions{
		Since:       after,
		State:       "all",
		ListOptions: githubApi.ListOptions{PerPage: 100},
	}

	for {

		issues, resp, err := handler.ghClient.Issues.ListByRepo(
			handler.ctx, owner, repo, listIssueOptions)

		if err != nil {
			return nil, err
		}

		allIssues = append(allIssues, issues...)

		if resp.NextPage == 0 {
			break
		}
		listIssueOptions.Page = resp.NextPage

	}

	for iIndex, issue := range allIssues {
		var pr *githubApi.PullRequest

		// base rfc object
		rfc := RFCUpdate{
			RFCID:   issue.GetID(),
			RFCType: RFCTypes[*project.ID],
			Issue: &github.Issue{
				Title:    issue.GetTitle(),
				URL:      issue.GetURL(),
				Number:   issue.GetNumber(),
				State:    issue.GetState(),
				Updated:  issue.GetUpdatedAt(),
				Created:  issue.GetCreatedAt(),
				Comments: issue.GetComments(),
				HtmlURL:  issue.GetHTMLURL(),
			},
		}

		if issue.IsPullRequest() {

			log.Printf("parsing %s. Progress %d/%d\n", url, iIndex+1, len(allIssues))

			pr, resp, err = handler.ghClient.PullRequests.Get(
				handler.ctx, owner, repo, issue.GetNumber(),
			)
			//github.RespMiddleware(resp)

			if err != nil {
				return nil, err
			}

			rfc.Issue.IsPR = issue.IsPullRequest()
			rfc.Issue.Merged = *pr.Merged

			if rfc.Issue.Merged {
				rfc.Issue.MergedAt = *pr.MergedAt
			}
		}

		// If is open and is not new (update) mark as update
		if rfc.Issue.Created != rfc.Issue.Updated &&
			!rfc.Issue.Merged &&
			rfc.Issue.State == "open" {
			rfc.Issue.IsUpdate = true
		}

		rfc.RFCNumber, err = GetRFCNumber(issue.GetTitle())
		if err != nil {
			return nil, err
		}

		if rfc.RFCNumber != -1 {

			switch rfc.RFCType {
			case RFCTypes[BIP]:
				rfc.RFCLink = fmt.Sprintf(BIPLink, rfc.RFCNumber)
			case RFCTypes[SLIP]:
				rfc.RFCLink = fmt.Sprintf(SLIPLink, rfc.RFCNumber)
			case RFCTypes[BOLT]:
				rfc.RFCLink = static.BoltMap[rfc.RFCNumber]
			}
		}

		post := &posts.Post{}
		post.Title = rfc.Issue.Title
		post.Link = rfc.Issue.URL
		post.Published = rfc.Issue.Created
		post.Updated = rfc.Issue.Updated
		post.JsonData = utils.StructToJsonMap(rfc)

		results = append(results, post)
	}

	return results, nil
}

func NewRFCHandler() FormatHandler {
	ctxb := context.Background()
	client := github.Auth(ctxb)

	return RFCHandler{
		ctx:      ctxb,
		ghClient: client,
	}
}

func GetRFCNumber(title string) (int, error) {

	// Detect BIP
	for _, re := range []*regexp.Regexp{ReBIP, ReBOLT, ReSLIP} {

		matches := re.FindStringSubmatch(strings.ToLower(title))

		if matches != nil {
			res, err := strconv.Atoi(matches[1])
			if err != nil {
				return -1, err
			}

			return res, nil
		}

	}

	return -1, nil
}
