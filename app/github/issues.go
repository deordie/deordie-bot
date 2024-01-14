package github

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

type Client struct {
	githubToken string
	owner       string
	repo        string
	issuesUrl   string
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

type createIssueRequest struct {
	Title  string   `json:"title"`
	Body   string   `json:"body"`
	Labels []string `json:"labels"`
}

type issue struct {
	Id      int64  `json:"id"`
	HtmlUrl string `json:"html_url"`
}

func wrapError(err error) error {
	return fmt.Errorf("error occurred during CreateIssue call: %w", err)
}

func NewClient(token string, githubRepo string) *Client {
	repoParts := strings.Split(githubRepo, "/")
	owner := repoParts[0]
	repo := repoParts[1]

	return &Client{
		githubToken: token,
		owner:       owner,
		repo:        repo,
		issuesUrl:   fmt.Sprintf("https://api.github.com/repos/%s/%s/issues", owner, repo),
	}
}

func (c *Client) CreateIssue(article *ArticleIssue) (string, error) {
	payload := newCreateIssueRequest(article)
	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return "", wrapError(err)
	}

	req, err := http.NewRequest("POST", c.issuesUrl, bytes.NewBuffer(jsonPayload))
	if err != nil {
		return "", wrapError(err)
	}

	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("Authorization", "Bearer "+c.githubToken)
	req.Header.Add("Content-Type", "application/json")

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", wrapError(err)
	}

	defer res.Body.Close()
	body, err := io.ReadAll(res.Body)
	if err != nil {
		return "", wrapError(err)
	}

	if res.StatusCode != 201 {
		return "", fmt.Errorf("non-successful HTTP status code in CreateIssue call: %d", res.StatusCode)
	}

	var iss issue
	err = json.Unmarshal(body, &iss)
	if err != nil {
		return "", wrapError(err)
	}

	return iss.HtmlUrl, nil
}

func newCreateIssueRequest(article *ArticleIssue) *createIssueRequest {
	title := article.Title
	if len(article.Author) > 0 {
		title = title + " / " + article.Author
	}

	body := fmt.Sprintf("__URL:__ %s\n\n__Review (1-2 sentences):__ %s\n\n__Created by:__ DE or DIE Bot :robot: on behalf of %s.", article.Url, article.Description, article.User)

	labels := make([]string, 0, len(article.Topics)+1)
	labels = append(labels, "level:"+article.Level)
	for _, topic := range article.Topics {
		labels = append(labels, "topic:"+topic)
	}

	return &createIssueRequest{
		Title:  title,
		Body:   body,
		Labels: labels,
	}
}
