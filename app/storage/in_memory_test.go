package storage

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestInMemoryStorage_SetAndGet(t *testing.T) {
	// Arrange
	storage := NewInMemoryStorage[string]()
	key := int64(1)
	value := "test"

	// Act
	storage.Set(key, value)
	result, ok := storage.Get(key)

	// Assert
	assert.True(t, ok, "Expected key to be found")
	assert.Equal(t, "test", result, "Expected value to match")
}

func TestInMemoryStorage_SetAndDelete(t *testing.T) {
	// Arrange
	storage := NewInMemoryStorage[string]()
	key := int64(1)
	value := "test"

	// Act
	storage.Set(key, value)
	storage.Delete(key)
	result, ok := storage.Get(key)

	// Assert
	assert.False(t, ok, "Expected key to not be found after deletion")
	assert.Equal(t, "", result, "Expected value to be empty after deletion")
}

func TestInMemoryStorage_GetNotFound(t *testing.T) {
	// Arrange
	storage := NewInMemoryStorage[string]()
	key := int64(1)

	// Act
	result, ok := storage.Get(key)

	// Assert
	assert.False(t, ok, "Expected key to not be found")
	assert.Equal(t, "", result, "Expected value to be empty")
}
