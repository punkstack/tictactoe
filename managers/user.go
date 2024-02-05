package managers

import (
	"errors"
	"sync"
	"tictactoe/models"
)

// UserManager manages user operations such as creation and retrieval.
type UserManager struct {
	users map[string]*models.User
	mu    sync.RWMutex // ensures thread-safe access to the users map
}

// NewUserManager creates a new UserManager instance.
func NewUserManager() *UserManager {
	return &UserManager{
		users: make(map[string]*models.User),
	}
}

// CreateUser creates a new user or retrieves an existing one based on the username.
func (m *UserManager) CreateUser(username, deviceID string) *models.User {
	m.mu.Lock()
	defer m.mu.Unlock()

	if user, exists := m.users[username]; exists {
		// Return existing user if found
		return user
	}

	// Create a new user since one doesn't exist
	newUser := models.NewUser(username, deviceID)
	m.users[username] = newUser
	return newUser
}

// GetUser retrieves an existing user by their username.
func (m *UserManager) GetUser(username string) (*models.User, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if user, exists := m.users[username]; exists {
		return user, nil
	}
	return nil, errors.New("user not found")
}

// UpdateUserStats updates the statistics for a user.
func (m *UserManager) UpdateUserStats(username string, won bool, draw bool) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	user, exists := m.users[username]
	if !exists {
		return errors.New("user not found")
	}

	if draw {
		user.UpdateStats(false, true)
	} else if won {
		user.UpdateStats(true, false)
	} else {
		user.UpdateStats(false, false)
	}
	return nil
}
