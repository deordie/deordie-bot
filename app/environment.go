package main

import (
	"errors"
	"fmt"
	"github.com/joho/godotenv"
	"io/fs"
	"log"
	"os"
	"sort"
	"strings"
)

type Environment struct {
	TelegramBotApiToken string
	PublicUrl           string
	RapidApiToken       string
	GitHubToken         string
	GitHubRepo          string
}

func LoadEnvironment() (*Environment, error) {
	err := godotenv.Load()
	if err != nil {
		var pathError *fs.PathError
		if !errors.As(err, &pathError) {
			log.Fatalf("Cannot parse .env file: %s", err.Error())
		}
	}

	envVars := map[string]string{"TELEGRAM_BOT_API_TOKEN": "", "PUBLIC_URL": "", "RAPID_API_TOKEN": "", "GITHUB_TOKEN": "", "GITHUB_REPO": ""}
	var missingVars []string
	for k := range envVars {
		envValue, ok := os.LookupEnv(k)
		if !ok || len(envValue) == 0 {
			missingVars = append(missingVars, k)
		}
		envVars[k] = envValue
	}

	if len(missingVars) > 0 {
		sort.Strings(missingVars)
		return nil, fmt.Errorf("missing environment variables: %s", strings.Join(missingVars, ", "))
	}

	return &Environment{
		TelegramBotApiToken: envVars["TELEGRAM_BOT_API_TOKEN"],
		PublicUrl:           envVars["PUBLIC_URL"],
		RapidApiToken:       envVars["RAPID_API_TOKEN"],
		GitHubToken:         envVars["GITHUB_TOKEN"],
		GitHubRepo:          envVars["GITHUB_REPO"],
	}, nil
}
