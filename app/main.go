package main

import (
	"errors"
	"fmt"
	"github.com/deordie/deordie-bot/app/github"
	"github.com/deordie/deordie-bot/app/rapidapi"
	"github.com/deordie/deordie-bot/app/telegram"
	"github.com/joho/godotenv"
	"io/fs"
	"log"
	"os"
)

type environment struct {
	TelegramBotApiToken string
	RapidApiToken       string
	GitHubToken         string
	GitHubRepo          string
}

func main() {
	env, err := loadEnvironment()
	if err != nil {
		log.Fatalf("Can't read environment: %s", err.Error())
	}

	rapidApiClient := rapidapi.NewClient(env.RapidApiToken)
	githubClient := github.NewClient(env.GitHubToken, env.GitHubRepo)
	bot, err := telegram.NewBot(env.TelegramBotApiToken, rapidApiClient, githubClient)
	if err != nil {
		log.Fatalf("Can't start bot: %s", err.Error())
	}

	bot.Start()
}

func loadEnvironment() (*environment, error) {
	err := godotenv.Load()
	if err != nil {
		var pathError *fs.PathError
		if !errors.As(err, &pathError) {
			log.Fatalf("Cannot parse .env file: %s", err.Error())
		}
	}

	envVars := map[string]string{"TELEGRAM_BOT_API_TOKEN": "", "RAPID_API_TOKEN": "", "GITHUB_TOKEN": "", "GITHUB_REPO": ""}
	for k := range envVars {
		envValue, ok := os.LookupEnv(k)
		if !ok || len(envValue) == 0 {
			return nil, fmt.Errorf("%s environment variable is required", k)
		}
		envVars[k] = envValue
	}

	return &environment{
		TelegramBotApiToken: envVars["TELEGRAM_BOT_API_TOKEN"],
		RapidApiToken:       envVars["RAPID_API_TOKEN"],
		GitHubToken:         envVars["GITHUB_TOKEN"],
		GitHubRepo:          envVars["GITHUB_REPO"],
	}, nil
}
