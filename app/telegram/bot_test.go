package telegram

import (
	"fmt"
	"github.com/deordie/deordie-bot/app/github"
	"github.com/deordie/deordie-bot/app/rapidapi"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	tele "gopkg.in/telebot.v3"
	"testing"
)

type MockTelegramBotContext struct {
	mock.Mock
	tele.Context
}

func (m *MockTelegramBotContext) Send(what interface{}, opts ...interface{}) error {
	args := m.Called(what, opts)
	return args.Error(0)
}

func (m *MockTelegramBotContext) Text() string {
	args := m.Called()
	return args.String(0)
}

func (m *MockTelegramBotContext) Sender() *tele.User {
	args := m.Called()
	return args.Get(0).(*tele.User)
}

type MockRapidAPIClient struct {
	mock.Mock
}

func (m *MockRapidAPIClient) ExtractArticle(articleUrl string) (*rapidapi.Article, error) {
	args := m.Called(articleUrl)
	return args.Get(0).(*rapidapi.Article), args.Error(1)
}

type MockGitHubClient struct {
	mock.Mock
}

func (m *MockGitHubClient) CreateIssue(article *github.ArticleIssue) (string, error) {
	args := m.Called(article)
	return args.String(0), args.Error(1)
}

func newTestBot(rapidApiClient *MockRapidAPIClient, githubClient *MockGitHubClient) *Bot {
	if rapidApiClient == nil {
		rapidApiClient = new(MockRapidAPIClient)
	}
	if githubClient == nil {
		githubClient = new(MockGitHubClient)
	}
	return &Bot{
		telebot:            &tele.Bot{},
		articleExtractor:   rapidApiClient,
		githubIssueCreator: githubClient,
		stateStorage:       NewStateStorage(),
	}
}

func TestStartHandler(t *testing.T) {
	bot := newTestBot(nil, nil)
	mockContext := new(MockTelegramBotContext)
	mockContext.On("Send", mock.Anything, mock.Anything).Return(nil)

	_ = bot.handleStart(mockContext)

	mockContext.AssertCalled(t, "Send", "Hello! I can help you to quickly propose an article for DE or DIE: Digest. Start with /newarticle command and follow the instructions.", mock.Anything)
}

func TestHelpHandler(t *testing.T) {
	bot := newTestBot(nil, nil)
	mockContext := new(MockTelegramBotContext)
	mockContext.On("Send", mock.Anything, mock.Anything).Return(nil)

	_ = bot.handleHelp(mockContext)

	mockContext.AssertCalled(t, "Send", "Supported commands:\n/newarticle - Propose an article for DE or DIE: Digest.", mock.Anything)
}

func TestNewArticleHandler(t *testing.T) {
	userId := int64(1000)
	bot := newTestBot(nil, nil)
	mockContext := new(MockTelegramBotContext)
	mockContext.On("Send", mock.Anything, mock.Anything).Return(nil)
	mockContext.On("Sender").Return(&tele.User{ID: userId})

	_ = bot.handleNewArticle(mockContext)

	_, ok := bot.stateStorage.Get(userId)
	assert.True(t, ok)
	mockContext.AssertCalled(t, "Send", "Step 1. Provide article URL. To abort the operation type \"cancel\".", mock.Anything)
}

func TestOnTextHandler_WhenUrlState(t *testing.T) {
	userId := int64(1001)
	bot := newTestBot(nil, nil)
	mockContext := new(MockTelegramBotContext)
	mockContext.On("Text").Return("https://example.com")
	mockContext.On("Send", mock.Anything, mock.Anything).Return(nil)
	mockContext.On("Sender").Return(&tele.User{ID: userId})
	bot.stateStorage.Set(userId, UserArticleState{UserId: userId})

	_ = bot.handleOnText(mockContext)

	expectedState := UserArticleState{
		UserId: userId,
		Url:    "https://example.com",
	}
	actualState, ok := bot.stateStorage.Get(userId)
	assert.True(t, ok)
	assert.Equal(t, expectedState, actualState)
	mockContext.AssertCalled(t, "Send", "Step 2. Provide article description as a plain text. To abort the operation type \"cancel\".", mock.Anything)
}

func TestOnTextHandler_WhenDescriptionState(t *testing.T) {
	userId := int64(1002)
	bot := newTestBot(nil, nil)
	mockContext := new(MockTelegramBotContext)
	mockContext.On("Text").Return("Nice article.")
	mockContext.On("Send", mock.Anything, mock.Anything).Return(nil)
	mockContext.On("Sender").Return(&tele.User{ID: userId})
	bot.stateStorage.Set(userId, UserArticleState{UserId: userId, Url: "https://example.com"})

	_ = bot.handleOnText(mockContext)

	expectedState := UserArticleState{
		UserId:      userId,
		Url:         "https://example.com",
		Description: "Nice article.",
	}
	actualState, ok := bot.stateStorage.Get(userId)
	assert.True(t, ok)
	assert.Equal(t, expectedState, actualState)
	// TODO: Cover keyboard with unit tests.
	mockContext.AssertCalled(t, "Send", "Step 3. Provide level. To abort the operation type \"cancel\".", mock.Anything)
}

func TestOnTextHandler_WhenLevelState(t *testing.T) {
	userId := int64(1003)
	bot := newTestBot(nil, nil)
	mockContext := new(MockTelegramBotContext)
	mockContext.On("Text").Return("advanced")
	mockContext.On("Send", mock.Anything, mock.Anything).Return(nil)
	mockContext.On("Sender").Return(&tele.User{ID: userId})
	bot.stateStorage.Set(userId, UserArticleState{UserId: userId, Url: "https://example.com", Description: "Nice article."})

	_ = bot.handleOnText(mockContext)

	expectedState := UserArticleState{
		UserId:      userId,
		Url:         "https://example.com",
		Description: "Nice article.",
		Level:       "advanced",
	}
	actualState, ok := bot.stateStorage.Get(userId)
	assert.True(t, ok)
	assert.Equal(t, expectedState, actualState)
	mockContext.AssertCalled(t, "Send", "Step 4. Provide topics as a comma separated list, e.g. \"streaming, storage-engine, kafka\" without quotes. To abort the operation type \"cancel\".", mock.Anything)
}

func TestOnTextHandler_WhenSuccessfullyCreateIssue(t *testing.T) {
	userId := int64(1004)
	mockRapidApi := new(MockRapidAPIClient)
	mockRapidApi.On("ExtractArticle", mock.Anything).Return(&rapidapi.Article{Title: "Article Title", Author: "Noname Blog", Url: "https://example.com/1"}, nil)
	mockGitHub := new(MockGitHubClient)
	mockGitHub.On("CreateIssue", mock.Anything).Return("https://github.com/deordie/deordie-digest/issues/1", nil)
	bot := newTestBot(mockRapidApi, mockGitHub)
	mockContext := new(MockTelegramBotContext)
	mockContext.On("Text").Return("topic1, topic2")
	mockContext.On("Send", mock.Anything, mock.Anything).Return(nil)
	mockContext.On("Sender").Return(&tele.User{ID: userId, Username: "nickname"})
	bot.stateStorage.Set(userId, UserArticleState{UserId: userId, Url: "https://example.com", Description: "Nice article.", Level: "advanced"})

	_ = bot.handleOnText(mockContext)

	expectedArticleIssue := &github.ArticleIssue{
		Url:         "https://example.com/1",
		Title:       "Article Title",
		Author:      "Noname Blog",
		Description: "Nice article.",
		Level:       "advanced",
		Topics:      []string{"topic1", "topic2"},
		User:        "https://t.me/nickname",
	}
	_, ok := bot.stateStorage.Get(userId)
	assert.False(t, ok)
	mockRapidApi.AssertCalled(t, "ExtractArticle", "https://example.com")
	mockGitHub.AssertCalled(t, "CreateIssue", expectedArticleIssue)
	mockContext.AssertCalled(t, "Send", "The article was added to the digest candidates! GitHub issue link: https://github.com/deordie/deordie-digest/issues/1", mock.Anything)
}

func TestOnTextHandler_WhenExtractArticleFailed(t *testing.T) {
	userId := int64(1005)
	mockRapidApi := new(MockRapidAPIClient)
	mockRapidApi.On("ExtractArticle", mock.Anything).Return(&rapidapi.Article{}, fmt.Errorf("extracting article error"))
	bot := newTestBot(mockRapidApi, nil)
	mockContext := new(MockTelegramBotContext)
	mockContext.On("Text").Return("topic1, topic2")
	mockContext.On("Send", mock.Anything, mock.Anything).Return(nil)
	mockContext.On("Sender").Return(&tele.User{ID: userId, Username: "nickname"})
	bot.stateStorage.Set(userId, UserArticleState{UserId: userId, Url: "https://example.com", Description: "Nice article.", Level: "advanced"})

	_ = bot.handleOnText(mockContext)

	_, ok := bot.stateStorage.Get(userId)
	assert.False(t, ok)
	mockRapidApi.AssertCalled(t, "ExtractArticle", "https://example.com")
	mockContext.AssertCalled(t, "Send", "Operation failed on fetching article.", mock.Anything)
}

func TestOnTextHandler_WhenCreateIssueFailed(t *testing.T) {
	userId := int64(1006)
	mockRapidApi := new(MockRapidAPIClient)
	mockRapidApi.On("ExtractArticle", mock.Anything).Return(&rapidapi.Article{Title: "Article Title", Author: "Noname Blog", Url: "https://example.com/1"}, nil)
	mockGitHub := new(MockGitHubClient)
	mockGitHub.On("CreateIssue", mock.Anything).Return("", fmt.Errorf("create issue error"))
	bot := newTestBot(mockRapidApi, mockGitHub)
	mockContext := new(MockTelegramBotContext)
	mockContext.On("Text").Return("topic1, topic2")
	mockContext.On("Send", mock.Anything, mock.Anything).Return(nil)
	mockContext.On("Sender").Return(&tele.User{ID: userId, Username: "nickname"})
	bot.stateStorage.Set(userId, UserArticleState{UserId: userId, Url: "https://example.com", Description: "Nice article.", Level: "advanced"})

	_ = bot.handleOnText(mockContext)

	expectedArticleIssue := &github.ArticleIssue{
		Url:         "https://example.com/1",
		Title:       "Article Title",
		Author:      "Noname Blog",
		Description: "Nice article.",
		Level:       "advanced",
		Topics:      []string{"topic1", "topic2"},
		User:        "https://t.me/nickname",
	}
	_, ok := bot.stateStorage.Get(userId)
	assert.False(t, ok)
	mockRapidApi.AssertCalled(t, "ExtractArticle", "https://example.com")
	mockGitHub.AssertCalled(t, "CreateIssue", expectedArticleIssue)
	mockContext.AssertCalled(t, "Send", "Operation failed on creating GitHub issue.", mock.Anything)
}

func TestOnTextHandler_WhenEmptyState(t *testing.T) {
	userId := int64(8888)
	bot := newTestBot(nil, nil)
	mockContext := new(MockTelegramBotContext)
	mockContext.On("Text").Return("Test")
	mockContext.On("Send", mock.Anything, mock.Anything).Return(nil)
	mockContext.On("Sender").Return(&tele.User{ID: userId})

	err := bot.handleOnText(mockContext)

	assert.Nil(t, err)
	mockContext.AssertNotCalled(t, "Send")
}

func TestOnTextHandler_WhenCancelled(t *testing.T) {
	userId := int64(9999)
	bot := newTestBot(nil, nil)
	mockContext := new(MockTelegramBotContext)
	mockContext.On("Text").Return("Cancel")
	mockContext.On("Send", mock.Anything, mock.Anything).Return(nil)
	mockContext.On("Sender").Return(&tele.User{ID: userId})
	bot.stateStorage.Set(userId, UserArticleState{UserId: userId})

	err := bot.handleOnText(mockContext)

	assert.Nil(t, err)
	mockContext.AssertCalled(t, "Send", "The operation was cancelled.", mock.Anything)
}
