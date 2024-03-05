package managers

import (
	"context"
	"sync"
	"tictactoe/models" // Adjust the import path based on your actual project structure
	"time"
)

// MatchmakingManager handles the matchmaking process.
type MatchmakingManager struct {
	mu          sync.Mutex
	matchmaking map[*models.User]chan *models.Game
}

// NewMatchmakingManager creates a new MatchmakingManager instance.
func NewMatchmakingManager() *MatchmakingManager {
	return &MatchmakingManager{
		matchmaking: make(map[*models.User]chan *models.Game),
	}
}

// RequestMatch handles a new matchmaking request.
func (m *MatchmakingManager) RequestMatch(ctx context.Context, user *models.User) (*models.Game, error) {
	matchChan := make(chan *models.Game, len(m.matchmaking)+1)

	// check and remove user from old game .... or  return current game

	m.mu.Lock()
	for opponent, oppChan := range m.matchmaking {
		// Found an opponent, remove them from the matchmaking pool and start a game
		delete(m.matchmaking, opponent)
		m.mu.Unlock()

		game := models.NewGame(user, opponent)

		// Notify both players' channels
		matchChan <- game
		oppChan <- game
		return game, nil
	}

	// No opponent found, add user to the matchmaking pool
	m.matchmaking[user] = matchChan
	m.mu.Unlock()

	select {
	case game := <-matchChan:
		// Match found
		return game, nil
	case <-time.After(120 * time.Second):
		// Matchmaking timeout
		m.mu.Lock()
		delete(m.matchmaking, user)
		m.mu.Unlock()
		return nil, nil // or an error indicating timeout
	case <-ctx.Done():
		// Context cancellation
		m.mu.Lock()
		delete(m.matchmaking, user)
		m.mu.Unlock()
		return nil, ctx.Err()
	}
}
