package telegram

import "github.com/deordie/deordie-bot/app/storage"

type UserArticleState struct {
	UserId      int64
	Url         string
	Description string
	Level       string
	Topics      []string
}

type StateStorage struct {
	storage.InMemoryStorage[UserArticleState]
}

func NewStateStorage() *StateStorage {
	return &StateStorage{
		InMemoryStorage: *storage.NewInMemoryStorage[UserArticleState](),
	}
}
