package managers

import (
	"errors"
	"fmt"
	"sync"
	"tictactoe/models" // Adjust the import path based on your actual project structure
	"time"
)

// GameManager manages game-related operations.
type GameManager struct {
	games map[string]*models.Game
	mu    sync.RWMutex // ensures thread-safe access to the games map
}

// NewGameManager creates a new instance of GameManager.
func NewGameManager() *GameManager {
	return &GameManager{
		games: make(map[string]*models.Game),
	}
}

// CreateGame initializes a new game with two players and adds it to the games map.
func (m *GameManager) CreateGame(player1, player2 *models.User) *models.Game {
	m.mu.Lock()
	defer m.mu.Unlock()

	gameID := fmt.Sprintf("game_%d", time.Now().UnixNano()) // Simple way to generate a unique game ID
	game := &models.Game{
		ID:          gameID,
		Players:     []*models.User{player1, player2},
		Board:       [3][3]string{},
		CurrentTurn: "X", // By default, the first player ("X") starts the game
		Status:      "in_progress",
	}

	m.games[gameID] = game
	return game
}

// UpdateGame processes a player's move and updates the game state.

func (m *GameManager) UpdateGame(gameID string, player *models.User, row, col int) (*models.Game, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	game, exists := m.games[gameID]
	if !exists {
		return nil, errors.New("game not found")
	}

	// Check if it's the player's turn
	// Assume player[0] is "X" and player[1] is "O"
	if !game.IsPlayerCurrent(player) {
		return nil, errors.New("not your turn")
	}

	if err := game.IsValidMove(row, col); err != nil {
		return nil, err
	}

	// Update the board
	game.UpdateBoard(row, col, game.CurrentTurn)

	// Check for a win or a draw
	if m.checkWin(game.Board, game.CurrentTurn) {
		game.UpdateWinState(player)
	} else if m.checkDraw(game.Board) {
		game.UpdateDrawState()
	} else {
		// Toggle the current turn
		game.CurrentTurn = m.toggleTurn(game.CurrentTurn)
	}

	return game, nil
}

func (m *GameManager) checkWin(board [3][3]string, playerSymbol string) bool {
	// Check rows, columns, and diagonals for a win
	for i := 0; i < 3; i++ {
		if board[i][0] == playerSymbol && board[i][1] == playerSymbol && board[i][2] == playerSymbol { // Check rows
			return true
		}
		if board[0][i] == playerSymbol && board[1][i] == playerSymbol && board[2][i] == playerSymbol { // Check columns
			return true
		}
	}
	// Check diagonals
	if board[0][0] == playerSymbol && board[1][1] == playerSymbol && board[2][2] == playerSymbol {
		return true
	}
	if board[0][2] == playerSymbol && board[1][1] == playerSymbol && board[2][0] == playerSymbol {
		return true
	}
	return false
}

func (m *GameManager) checkDraw(board [3][3]string) bool {
	for i := 0; i < 3; i++ {
		for j := 0; j < 3; j++ {
			if board[i][j] == "" {
				return false // If any cell is empty, the game is not a draw
			}
		}
	}
	return true // No empty cells, game is a draw
}

func (m *GameManager) toggleTurn(currentTurn string) string {
	// Toggle the current turn and return the new turn
	if currentTurn == "X" {
		return "O"
	}
	return "X"
}

// GetGame retrieves a game by its ID.
func (m *GameManager) GetGame(gameID string) (*models.Game, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	game, exists := m.games[gameID]
	if !exists {
		return nil, fmt.Errorf("game with ID %s not found", gameID)
	}

	return game, nil
}

// EndGame marks a game as completed and sets the winner.
func (m *GameManager) EndGame(gameID, winner string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	game, exists := m.games[gameID]
	if !exists {
		return fmt.Errorf("game with ID %s not found", gameID)
	}

	game.Status = "completed"
	game.Winner = winner
	return nil
}
