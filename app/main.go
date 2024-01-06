package main

import (
	"github.com/deordie/deordie-bot/app/github"
	"github.com/deordie/deordie-bot/app/rapidapi"
	"github.com/deordie/deordie-bot/app/telegram"
	"log"
)

func main() {
	env, err := LoadEnvironment()
	if err != nil {
		log.Fatalf("Can't read environment: %s", err.Error())
	}

	rapidApiClient := rapidapi.NewClient(env.RapidApiToken)
	githubClient := github.NewClient(env.GitHubToken, env.GitHubRepo)
	bot, err := telegram.NewBot(env.TelegramBotApiToken, rapidApiClient, githubClient, env.PublicUrl)
	if err != nil {
		log.Fatalf("Can't start bot: %s", err.Error())
	}

	bot.Start()
}
