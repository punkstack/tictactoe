package models

import (
	"github.com/gorilla/websocket"
	"sync"
)

// User represents a player or user in the system.
type User struct {
	Username    string          // Unique identifier for the user
	DeviceID    string          // Device identifier for the user, if applicable
	Conn        *websocket.Conn // WebSocket connection for real-time communication
	CurrentGame *Game           // Pointer to the current game the user is part of, if any
	Stats       UserStats       // User's game statistics
	mu          sync.Mutex      // Mutex for synchronizing writes
}

// UserStats holds the statistics related to game outcomes for the user.
type UserStats struct {
	Wins   int // Number of games won by the user
	Losses int // Number of games lost by the user
	Draws  int // Number of games that ended in a draw
}

// NewUser initializes a new User instance.
func NewUser(username, deviceID string) *User {
	return &User{
		Username: username,
		DeviceID: deviceID,
		Stats:    UserStats{},
	}
}

func (u *User) SetConnection(conn *websocket.Conn) {
	u.Conn = conn
}

// UpdateStats updates the user's game statistics based on the game outcome.
func (u *User) UpdateStats(won bool, draw bool) {
	if draw {
		u.Stats.Draws++
		return
	}
	if won {
		u.Stats.Wins++
	} else {
		u.Stats.Losses++
	}
}

func (u *User) SendMessage(msg []byte) error {
	u.mu.Lock()
	defer u.mu.Unlock()
	return u.Conn.WriteMessage(websocket.TextMessage, msg)
}
