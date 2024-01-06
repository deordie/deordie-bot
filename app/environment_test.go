package main

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLoadEnvironment(t *testing.T) {
	err := os.Setenv("TELEGRAM_BOT_API_TOKEN", "telegram_token")
	assert.NoError(t, err)
	err = os.Setenv("RAPID_API_TOKEN", "rapid_api_token")
	assert.NoError(t, err)
	err = os.Setenv("GITHUB_TOKEN", "github_token")
	assert.NoError(t, err)
	err = os.Setenv("GITHUB_REPO", "github/repo")
	assert.NoError(t, err)
	err = os.Setenv("PUBLIC_URL", "https://example.com")
	assert.NoError(t, err)

	defer func() {
		_ = os.Unsetenv("TELEGRAM_BOT_API_TOKEN")
		_ = os.Unsetenv("RAPID_API_TOKEN")
		_ = os.Unsetenv("GITHUB_TOKEN")
		_ = os.Unsetenv("GITHUB_REPO")
		_ = os.Unsetenv("PUBLIC_URL")
	}()

	env, err := LoadEnvironment()
	assert.NoError(t, err)
	assert.NotNil(t, env)

	assert.Equal(t, "telegram_token", env.TelegramBotApiToken)
	assert.Equal(t, "rapid_api_token", env.RapidApiToken)
	assert.Equal(t, "github_token", env.GitHubToken)
	assert.Equal(t, "github/repo", env.GitHubRepo)
	assert.Equal(t, "https://example.com", env.PublicUrl)
}

func TestLoadEnvironmentMissingToken(t *testing.T) {
	_ = os.Unsetenv("TELEGRAM_BOT_API_TOKEN")
	_ = os.Unsetenv("RAPID_API_TOKEN")
	_ = os.Unsetenv("GITHUB_TOKEN")
	_ = os.Unsetenv("GITHUB_REPO")
	_ = os.Unsetenv("PUBLIC_URL")

	_, err := LoadEnvironment()
	assert.Error(t, err)
	assert.Equal(t, err.Error(), "missing environment variables: GITHUB_REPO, GITHUB_TOKEN, PUBLIC_URL, RAPID_API_TOKEN, TELEGRAM_BOT_API_TOKEN")
}
