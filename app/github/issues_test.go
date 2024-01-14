package github

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCreateIssue_Success(t *testing.T) {
	// Arrange
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify request headers
		assert.Equal(t, "Bearer FAKE_GITHUB_TOKEN", r.Header.Get("Authorization"), "unexpected Authorization header")
		assert.Equal(t, "application/vnd.github+json", r.Header.Get("Accept"), "unexpected Accept header")
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"), "unexpected Content-Type header")

		body, _ := io.ReadAll(r.Body)
		defer r.Body.Close()

		var req createIssueRequest
		_ = json.Unmarshal(body, &req)

		expectedRequest := createIssueRequest{
			Title:  "Sample Title / John Doe",
			Body:   "__URL:__ https://example.com\n\n__Review (1-2 sentences):__ Sample description\n\n__Created by:__ DE or DIE Bot :robot: on behalf of user123.",
			Labels: []string{"level:beginner", "topic:topic1", "topic:topic2"},
		}
		assert.Equal(t, expectedRequest, req)

		w.Header().Set("Content-Type", "application/vnd.github+json")
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte(`{"id": 1, "html_url": "https://github.com/owner/repo/issues/1"}`))
	}))
	defer mockServer.Close()

	client := &Client{
		githubToken: "FAKE_GITHUB_TOKEN",
		owner:       "owner",
		repo:        "repo",
		issuesUrl:   mockServer.URL,
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
}

func TestCreateIssue_NonSuccessHttpStatus(t *testing.T) {
	// Arrange
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer mockServer.Close()

	client := &Client{
		githubToken: "FAKE_GITHUB_TOKEN",
		owner:       "owner",
		repo:        "repo",
		issuesUrl:   mockServer.URL,
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
	assert.EqualError(t, err, "non-successful HTTP status code in CreateIssue call: 404", "unexpected error message")
}

func TestCreateIssue_Failure(t *testing.T) {
	// Arrange
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/vnd.github+json")
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte(`malformed JSON`))
	}))
	defer mockServer.Close()

	client := &Client{
		githubToken: "FAKE_GITHUB_TOKEN",
		owner:       "owner",
		repo:        "repo",
		issuesUrl:   mockServer.URL,
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
	assert.EqualError(t, err, "error occurred during CreateIssue call: invalid character 'm' looking for beginning of value", "unexpected error message")
}
