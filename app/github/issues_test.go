package github

import (
	"context"
	"fmt"
	"net/http"
	"testing"

	"github.com/google/go-github/v57/github"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type issuesServiceMock struct {
	mock.Mock
}

func (m *issuesServiceMock) Create(ctx context.Context, owner string, repo string, req *github.IssueRequest) (*github.Issue, *github.Response, error) {
	args := m.Called(ctx, owner, repo, req)
	return args.Get(0).(*github.Issue), args.Get(1).(*github.Response), args.Error(2)
}

func TestCreateIssue_Success(t *testing.T) {
	// Arrange
	title := "Sample Title / John Doe"
	body := "__URL:__ https://example.com\n\n__Review (1-2 sentences):__ Sample description\n\n__Created by:__ DE or DIE Bot :robot: on behalf of user123."
	labels := []string{"level:beginner", "topic:topic1", "topic:topic2"}
	request := &github.IssueRequest{
		Title:  &title,
		Body:   &body,
		Labels: &labels,
	}
	response := &github.Response{Response: &http.Response{StatusCode: http.StatusCreated}}
	issue := &github.Issue{HTMLURL: github.String("https://github.com/owner/repo/issues/1")}
	mockIssuesService := new(issuesServiceMock)
	mockIssuesService.On("Create", mock.Anything, mock.Anything, mock.Anything, request).Return(issue, response, nil)

	client := &Client{
		issueCreator: mockIssuesService,
		owner:        "owner",
		repo:         "repo",
	}

	article := &ArticleIssue{
		Url:         "https://example.com",
		Title:       "Sample Title",
		Author:      "John Doe",
		Description: "Sample description",
		Level:       "beginner",
		Topics:      []string{"topic1", "topic2"},
		User:        "user123",
	}

	// Act
	url, err := client.CreateIssue(article)
	assert.Nil(t, err, "unexpected error")

	// Assert
	assert.Equal(t, "https://github.com/owner/repo/issues/1", url, "unexpected issue URL")
	mockIssuesService.AssertExpectations(t)
}

func TestCreateIssue_NonSuccessHttpStatus(t *testing.T) {
	// Arrange
	response := &github.Response{Response: &http.Response{StatusCode: http.StatusTooManyRequests}}
	mockIssuesService := new(issuesServiceMock)
	mockIssuesService.On("Create", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(&github.Issue{}, response, nil)

	client := &Client{
		issueCreator: mockIssuesService,
		owner:        "owner",
		repo:         "repo",
	}

	article := &ArticleIssue{
		Url:         "https://example.com",
		Title:       "Sample Title",
		Author:      "John Doe",
		Description: "Sample description",
		Level:       "beginner",
		Topics:      []string{"topic1", "topic2"},
		User:        "user123",
	}

	// Act
	_, err := client.CreateIssue(article)

	// Assert
	assert.NotNil(t, err, "expected non-nil error")
	assert.EqualError(t, err, "non-successful HTTP status code in CreateIssue call: 429", "unexpected error message")
}

func TestCreateIssue_Failure(t *testing.T) {
	// Arrange
	serviceError := fmt.Errorf("error creating issue")
	mockIssuesService := new(issuesServiceMock)
	mockIssuesService.On("Create", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(&github.Issue{}, &github.Response{}, serviceError)

	client := &Client{
		issueCreator: mockIssuesService,
		owner:        "owner",
		repo:         "repo",
	}

	article := &ArticleIssue{
		Url:         "https://example.com",
		Title:       "Sample Title",
		Author:      "John Doe",
		Description: "Sample description",
		Level:       "beginner",
		Topics:      []string{"topic1", "topic2"},
		User:        "user123",
	}

	// Act
	_, err := client.CreateIssue(article)

	// Assert
	assert.NotNil(t, err, "expected non-nil error")
	assert.EqualError(t, err, "error occurred during CreateIssue call: error creating issue", "unexpected error message")
}
