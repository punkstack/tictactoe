package models

import (
	"errors"
	"sync"
	"tictactoe/utils"
)

// Game represents a single game session between two players.
type Game struct {
	ID          string       // Unique identifier for the game
	Players     []*User      // Slice of pointers to User structs representing the players
	Board       [3][3]string // 3x3 board for Tic-Tac-Toe, each cell can be "X", "O", or empty
	CurrentTurn string       // Indicates whose turn it is - "X" or "O"
	Status      string       // Current status of the game, e.g., "waiting", "in_progress", "completed"
	Winner      string       // Winner of the game, if applicable - "X", "O", or "draw"
	mu          sync.Mutex
}

// NewGame initializes a new Game instance with two players.
func NewGame(player1, player2 *User) *Game {
	gameID := utils.GenerateGameID() // Assuming generateGameID is a function that generates a unique game ID
	return &Game{
		ID:          gameID,
		Players:     []*User{player1, player2},
		Board:       [3][3]string{},
		CurrentTurn: "X", // By default, the first player ("X") starts the game
		Status:      utils.GameStateInProgress,
	}
}

func (g *Game) IsPlayerCurrent(player *User) bool {
	var playerSymbol string
	if g.Players[0] == player {
		playerSymbol = "X"
	} else if g.Players[1] == player {
		playerSymbol = "O"
	} else {
		return false
	}

	return playerSymbol == g.CurrentTurn
}

func (g *Game) IsValidMove(row, col int) error {

	if row < 0 || row >= 3 || col < 0 || col >= 3 {
		return errors.New("move out of bounds")
	}

	if g.Board[row][col] != "" {
		return errors.New("cell already occupied")
	}

	return nil
}

func (g *Game) UpdateWinState(player *User) {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.Winner = player.Username
	g.Status = utils.GameStateCompleted
}

func (g *Game) UpdateDrawState() {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.Status = utils.GameStateCompleted
	g.Winner = utils.GameStateDraw
}

func (g *Game) UpdateBoard(row, col int, turn string) {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.Board[row][col] = turn
}
