package rapidapi

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestExtractArticle_Success(t *testing.T) {
	// Arrange
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify request headers
		assert.Equal(t, "FAKE_API_KEY", r.Header.Get("X-RapidAPI-Key"), "unexpected X-RapidAPI-Key header")
		assert.Equal(t, "full-text-rss.p.rapidapi.com", r.Header.Get("X-RapidAPI-Host"), "unexpected X-RapidAPI-Host header")
		body, _ := io.ReadAll(r.Body)
		defer r.Body.Close()
		assert.Equal(t, "url=https://example.com&xss=1&lang=2&links=preserve&content=0", string(body))

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"title": "Sample Title", "date": "2022-01-01T12:00:00Z", "author": "John Doe", "language": "en", "url": "https://example.com", "effective_url": "https://example.com", "domain": "example.com"}`))
	}))
	defer mockServer.Close()

	client := Client{
		apiKey:            "FAKE_API_KEY",
		fullTextRssApiUrl: mockServer.URL,
	}

	// Act
	article, err := client.ExtractArticle("https://example.com")

	// Assert
	assert.Nil(t, err, "unexpected error")
	expectedArticle := &Article{
		Title:        "Sample Title",
		Date:         time.Date(2022, 1, 1, 12, 0, 0, 0, time.UTC),
		Author:       "John Doe",
		Language:     "en",
		Url:          "https://example.com",
		EffectiveUrl: "https://example.com",
		Domain:       "example.com",
	}
	assert.Equal(t, expectedArticle, article, "unexpected extracted article")
}

func TestExtractArticle_NonSuccessHTTPStatus(t *testing.T) {
	// Arrange
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer mockServer.Close()

	client := Client{
		apiKey:            "FAKE_API_KEY",
		fullTextRssApiUrl: mockServer.URL,
	}

	// Act
	_, err := client.ExtractArticle("https://example.com")

	// Assert
	assert.NotNil(t, err, "expected non-nil error")
	assert.EqualError(t, err, "non-successful HTTP status code in ExtractArticle call: 404", "unexpected error message")
}

func TestExtractArticle_Failure(t *testing.T) {
	// Arrange
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`malformed JSON`))
	}))
	defer mockServer.Close()

	client := Client{
		apiKey:            "FAKE_API_KEY",
		fullTextRssApiUrl: mockServer.URL,
	}

	// Act
	_, err := client.ExtractArticle("https://example.com")

	// Assert
	assert.NotNil(t, err, "expected non-nil error")
	assert.EqualError(t, err, "error occurred during ExtractArticle call: invalid character 'm' looking for beginning of value", "unexpected error message")
}
