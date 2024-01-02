package storage

import (
	"sync"
)

type InMemoryStorage[T any] struct {
	mu sync.RWMutex
	m  map[int64]T
}

func NewInMemoryStorage[T any]() *InMemoryStorage[T] {
	return &InMemoryStorage[T]{
		m: make(map[int64]T),
	}
}

func (s *InMemoryStorage[T]) Set(key int64, value T) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.m[key] = value
}

func (s *InMemoryStorage[T]) Get(key int64) (T, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	value, ok := s.m[key]
	return value, ok
}

func (s *InMemoryStorage[T]) Delete(key int64) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.m, key)
}
