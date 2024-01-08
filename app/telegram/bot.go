package telegram

import (
	"fmt"
	"github.com/deordie/deordie-bot/app/github"
	"github.com/deordie/deordie-bot/app/rapidapi"
	tele "gopkg.in/telebot.v3"
	"log"
	"net/url"
	"strings"
)

const (
	startCommand      = "/start"
	newArticleCommand = "/newarticle"
	helpCommand       = "/help"
)

type articleExtractor interface {
	ExtractArticle(articleUrl string) (*rapidapi.Article, error)
}

type githubIssueCreator interface {
	CreateIssue(article *github.ArticleIssue) (string, error)
}

type Bot struct {
	telebot            *tele.Bot
	articleExtractor   articleExtractor
	githubIssueCreator githubIssueCreator
	stateStorage       *StateStorage
}

func NewBot(token string, rapidApiClient *rapidapi.Client, githubClient *github.Client, publicUrl string) (*Bot, error) {
	pref := tele.Settings{
		Token:  token,
		Poller: &tele.Webhook{Listen: ":8080", Endpoint: &tele.WebhookEndpoint{PublicURL: publicUrl}},
	}
	telebot, err := tele.NewBot(pref)
	if err != nil {
		return nil, fmt.Errorf("error occured during Telgram bot creation: %w", err)
	}

	log.Printf("The bot is configured as a Webhook with public URL: %s and listens on port 8080\n", publicUrl)

	return &Bot{
		telebot:            telebot,
		articleExtractor:   rapidApiClient,
		githubIssueCreator: githubClient,
		stateStorage:       NewStateStorage(),
	}, nil
}

func (b *Bot) Start() {
	b.telebot.Handle(startCommand, b.handleStart)
	b.telebot.Handle(newArticleCommand, b.handleNewArticle)
	b.telebot.Handle(helpCommand, b.handleHelp)
	b.telebot.Handle(tele.OnText, b.handleOnText)

	log.Printf("The bot is running...\n")
	b.telebot.Start()
}

func (b *Bot) handleStart(ctx tele.Context) error {
	var startText = fmt.Sprintf("Hello! I can help you to quickly propose an article for DE or DIE: Digest. Start with %s command and follow the instructions.", newArticleCommand)
	return ctx.Send(startText, tele.RemoveKeyboard)
}

func (b *Bot) handleHelp(ctx tele.Context) error {
	helpText := fmt.Sprintf(`Supported commands:
%s - Propose an article for DE or DIE: Digest.`, newArticleCommand)
	return ctx.Send(helpText, tele.RemoveKeyboard)
}

func (b *Bot) handleNewArticle(ctx tele.Context) error {
	b.stateStorage.Set(ctx.Sender().ID, UserArticleState{UserId: ctx.Sender().ID})
	return ctx.Send("Step 1. Provide article URL. To abort the operation type \"cancel\".", tele.RemoveKeyboard)
}

func (b *Bot) handleOnText(ctx tele.Context) error {
	userId := ctx.Sender().ID
	state, ok := b.stateStorage.Get(userId)
	if !ok {
		return nil
	}

	if strings.ToLower(ctx.Text()) == "cancel" {
		b.stateStorage.Delete(userId)
		return ctx.Send("The operation was cancelled.", tele.RemoveKeyboard)
	}

	if state.Url == "" {
		validatedUrl, err := url.ParseRequestURI(ctx.Text())
		if err != nil {
			return ctx.Send("Provided input is not a valid URL, please fix the URL or abort the operation by typing \"cancel\".")
		}

		state.Url = validatedUrl.String()
		b.stateStorage.Set(userId, state)
		return ctx.Send("Step 2. Provide article description as a plain text. To abort the operation type \"cancel\".")
	}

	if state.Description == "" {
		state.Description = ctx.Text()
		b.stateStorage.Set(userId, state)
		return ctx.Send("Step 3. Provide level. To abort the operation type \"cancel\".", getLevelKeyboard())
	}

	if state.Level == "" {
		state.Level = ctx.Text()
		b.stateStorage.Set(userId, state)
		return ctx.Send("Step 4. Provide topics as a comma separated list, e.g. \"streaming, storage-engine, kafka\" without quotes. To abort the operation type \"cancel\".", tele.RemoveKeyboard)
	}

	if len(state.Topics) == 0 {
		topics := strings.Split(ctx.Text(), ",")
		for i := range topics {
			topics[i] = strings.TrimSpace(topics[i])
		}
		state.Topics = topics
	}

	b.stateStorage.Delete(userId)

	article, err := b.articleExtractor.ExtractArticle(state.Url)
	if err != nil {
		log.Printf("Failed to extract article: %s", err.Error())
		return ctx.Send("Operation failed on fetching article.")
	}

	articleIssue := newArticleIssue(ctx.Sender().Username, article, &state)
	issueUrl, err := b.githubIssueCreator.CreateIssue(articleIssue)
	if err != nil {
		log.Printf("Failed to create GitHub issue: %s.\nExtracted article is:\n%v\n", err.Error(), article)
		return ctx.Send("Operation failed on creating GitHub issue.")
	}

	return ctx.Send("The article was added to the digest candidates! GitHub issue link: " + issueUrl)
}

func getLevelKeyboard() *tele.ReplyMarkup {
	keyboard := &tele.ReplyMarkup{ResizeKeyboard: true, OneTimeKeyboard: true}
	btnBeginner := keyboard.Text("beginner")
	btnMedium := keyboard.Text("medium")
	btnAdvanced := keyboard.Text("advanced")
	keyboard.Reply(keyboard.Row(btnBeginner, btnMedium, btnAdvanced))
	return keyboard
}

func newArticleIssue(user string, article *rapidapi.Article, state *UserArticleState) *github.ArticleIssue {
	return &github.ArticleIssue{
		Url:         article.Url,
		Title:       article.Title,
		Author:      article.Author,
		Description: state.Description,
		Level:       state.Level,
		Topics:      state.Topics,
		User:        fmt.Sprintf("https://t.me/%s", user),
	}
}
