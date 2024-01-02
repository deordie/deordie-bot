package github

import (
	"context"
	"fmt"
	"github.com/google/go-github/v57/github"
	"net/http"
	"strings"
)

type issueCreator interface {
	Create(ctx context.Context, owner string, repo string, issue *github.IssueRequest) (*github.Issue, *github.Response, error)
}

type Client struct {
	issueCreator issueCreator
	owner        string
	repo         string
}

type ArticleIssue struct {
	Url         string
	Title       string
	Author      string
	Description string
	Level       string
	Topics      []string
	User        string
}

func NewClient(token string, githubRepo string) *Client {
	client := github.NewClient(http.DefaultClient).WithAuthToken(token)
	repoParts := strings.Split(githubRepo, "/")

	return &Client{
		issueCreator: client.Issues,
		owner:        repoParts[0],
		repo:         repoParts[1],
	}
}

func (c *Client) CreateIssue(article *ArticleIssue) (string, error) {
	ctx := context.Background()
	req := createIssueRequest(article)
	issue, res, err := c.issueCreator.Create(ctx, c.owner, c.repo, req)
	if err != nil {
		return "", fmt.Errorf("error occurred during CreateIssue call: %w", err)
	}

	if res.StatusCode != 201 {
		return "", fmt.Errorf("non-successful HTTP status code in CreateIssue call: %d", res.StatusCode)
	}

	return *issue.HTMLURL, nil
}

func createIssueRequest(article *ArticleIssue) *github.IssueRequest {
	title := article.Title
	if len(article.Author) > 0 {
		title = title + " / " + article.Author
	}
	body := fmt.Sprintf("__URL:__ %s\n\n__Review (1-2 sentences):__ %s\n\n__Created by:__ DE or DIE Bot :robot: on behalf of %s.", article.Url, article.Description, article.User)
	labels := createLabels(article)
	return &github.IssueRequest{
		Title:  &title,
		Body:   &body,
		Labels: &labels,
	}
}

func createLabels(article *ArticleIssue) []string {
	labels := make([]string, 0, len(article.Topics)+1)
	labels = append(labels, "level:"+article.Level)

	for _, topic := range article.Topics {
		labels = append(labels, "topic:"+topic)
	}

	return labels
}
