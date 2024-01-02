package rapidapi

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

type Client struct {
	apiKey            string
	fullTextRssApiUrl string
}

type Article struct {
	Title        string    `json:"title"`
	Date         time.Time `json:"date"`
	Author       string    `json:"author"`
	Language     string    `json:"language"`
	Url          string    `json:"url"`
	EffectiveUrl string    `json:"effective_url"`
	Domain       string    `json:"domain"`
}

func NewClient(apiKey string) *Client {
	return &Client{
		apiKey:            apiKey,
		fullTextRssApiUrl: "https://full-text-rss.p.rapidapi.com/extract.php",
	}
}

func (c *Client) ExtractArticle(articleUrl string) (*Article, error) {
	payload := strings.NewReader(fmt.Sprintf("url=%s&xss=1&lang=2&links=preserve&content=0", articleUrl))

	req, err := http.NewRequest("POST", c.fullTextRssApiUrl, payload)
	if err != nil {
		return nil, fmt.Errorf("error occured during ExtractArticle call: %w", err)
	}

	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Add("X-RapidAPI-Key", c.apiKey)
	req.Header.Add("X-RapidAPI-Host", "full-text-rss.p.rapidapi.com")

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error occured during ExtractArticle call: %w", err)
	}

	defer res.Body.Close()
	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("error occurred during ExtractArticle call: %w", err)
	}

	if res.StatusCode != 200 {
		return nil, fmt.Errorf("non-successful HTTP status code in ExtractArticle call: %d", res.StatusCode)
	}

	var article Article
	err = json.Unmarshal(body, &article)
	if err != nil {
		return nil, fmt.Errorf("error occurred during ExtractArticle call: %w", err)
	}

	return &article, nil
}
