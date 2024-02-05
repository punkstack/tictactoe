package utils

import (
	"github.com/google/uuid"
)

// generateGameID creates a unique identifier for a game session.
func GenerateGameID() string {
	return uuid.New().String()
}
