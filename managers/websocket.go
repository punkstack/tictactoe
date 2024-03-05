package managers

import (
	"context"
	"encoding/json"
	"github.com/gorilla/websocket"
	"net/http"
	"sync"
	"tictactoe/models" // Adjust this import path to match your project's structure
	"tictactoe/utils"
)

// WebSocketManager manages WebSocket connections and messaging.
type WebSocketManager struct {
	clients            map[*websocket.Conn]*models.User // Maps connections to users
	userManager        *UserManager
	gameManager        *GameManager
	upgrader           websocket.Upgrader
	register           chan *websocket.Conn
	unregister         chan *websocket.Conn
	matchmakingManager *MatchmakingManager
	mu                 sync.Mutex // Protects the clients map
}

// NewWebSocketManager creates a new instance and starts its main loop.
func NewWebSocketManager(userManager *UserManager, gameManager *GameManager, matchmakingManager *MatchmakingManager) *WebSocketManager {
	wsm := &WebSocketManager{
		clients:     make(map[*websocket.Conn]*models.User),
		userManager: userManager,
		gameManager: gameManager,
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				return true // Bypass the origin check
			},
		},
		register:           make(chan *websocket.Conn),
		matchmakingManager: matchmakingManager,
		unregister:         make(chan *websocket.Conn),
	}
	go wsm.run()
	return wsm
}

// run listens for register/unregister requests and manages clients.
func (wsm *WebSocketManager) run() {
	for {
		select {
		case conn := <-wsm.register:
			// Register a new client
			wsm.mu.Lock()
			wsm.clients[conn] = nil // Initially, no user is associated with the connection
			wsm.mu.Unlock()

		case conn := <-wsm.unregister:
			// Unregister a client
			wsm.mu.Lock()
			if _, ok := wsm.clients[conn]; ok {
				delete(wsm.clients, conn)
				conn.Close() // Close the WebSocket connection
			}
			wsm.mu.Unlock()
		}
	}
}

func (wsm *WebSocketManager) HandleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := wsm.upgrader.Upgrade(w, r, nil) // Upgrade the HTTP connection to a WebSocket connection
	if err != nil {
		// Log the error or send an HTTP error response
		http.Error(w, "Could not upgrade to WebSocket", http.StatusInternalServerError)
		return
	}

	// Register the new WebSocket connection with the manager
	wsm.register <- conn

	// Start a goroutine to handle messages from this connection
	go wsm.handleMessages(conn)
}

// handleMessages reads and processes messages from the connection.
func (wsm *WebSocketManager) handleMessages(conn *websocket.Conn) {
	defer func() {
		user := wsm.clients[conn]
		if user != nil {
			wsm.userManager.UpdateUserStats(user.Username, false, false) // Update stats for disconnects, if necessary
		}
		wsm.unregister <- conn
		conn.Close()
	}()

	for {
		_, message, err := conn.ReadMessage()
		if err != nil {
			wsm.sendError(conn, "Failed to read message")
			break
		}

		var basePacket models.BasePacket
		if err := json.Unmarshal(message, &basePacket); err != nil {
			wsm.sendError(conn, "Invalid packet format")
			continue
		}

		switch basePacket.Type {
		case utils.ConnectPacketType:
			var packet models.ConnectPacket
			if err := json.Unmarshal(message, &packet); err != nil {
				wsm.sendError(conn, "Invalid connect packet format")
				continue
			}
			user := wsm.userManager.CreateUser(packet.Username, packet.DeviceID)
			wsm.clients[conn] = user
			user.SetConnection(conn)
			wsm.sendUserStats(conn, user)

		case utils.PlayPacketType:
			var packet models.PlayPacket
			if err := json.Unmarshal(message, &packet); err != nil {
				wsm.sendError(conn, "Invalid play packet format")
				continue
			}
			var err error
			user, err := wsm.userManager.GetUser(packet.Username)
			if err != nil {
				wsm.sendError(conn, "User not registered")
				continue
			}

			for _, game := range wsm.gameManager.games {
				if game.Players[0].DeviceID == user.DeviceID || game.Players[1].DeviceID == user.DeviceID {
					wsm.notifyGameUpdate(game)
					continue
				}
			}
			// Handle matchmaking and game initiation
			game, err := wsm.matchmakingManager.RequestMatch(context.Background(), user)
			if err != nil {
				wsm.sendError(conn, "Error in matchmaking")
				continue
			}
			if game != nil {
				wsm.gameManager.mu.Lock()
				wsm.gameManager.games[game.ID] = game
				// Game found, notify both players
				wsm.notifyGameStart(game)
				wsm.gameManager.mu.Unlock()
			} else {
				// No match found within timeout, notify player
				wsm.sendNoMatchFound(conn)
			}

		case utils.MovePacketType:
			var packet models.MovePacket
			if err := json.Unmarshal(message, &packet); err != nil {
				wsm.sendError(conn, "Invalid move packet format")
				continue
			}
			user, ok := wsm.clients[conn]
			if !ok {
				wsm.sendError(conn, "User not registered")
				continue
			}
			// Validate and process the move, update game state
			game, err := wsm.gameManager.UpdateGame(packet.GameID, user, packet.Row, packet.Col)
			if err != nil {
				wsm.sendError(conn, "Invalid move or not your turn")
				continue
			}

			if game != nil {
				// Game found, notify both players
				wsm.notifyGameUpdate(game)
			} else {
				// No match found within timeout, notify player
				wsm.sendNoMatchFound(conn)
			}

		// Handle other packet types as necessary

		default:
			wsm.sendError(conn, "Unknown packet type")
		}
	}
}

// sendUserStats sends the user's stats back to the client.
func (wsm *WebSocketManager) sendUserStats(conn *websocket.Conn, user *models.User) {
	userStatsPacket := models.GameUpdatePacket{
		BasePacket:  models.BasePacket{Type: utils.UserStatsPacketType},
		GameID:      "",             // No game ID needed for user stats
		Board:       [3][3]string{}, // Empty for user stats
		CurrentTurn: false,          // Empty for user stats
		Winner:      "",             // Empty for user stats, could include additional stats fields as needed
		Status:      "started",
	}
	msg, err := json.Marshal(userStatsPacket)
	if err != nil {
		// Log error, handle failure to marshal packet
		return
	}
	conn.WriteMessage(websocket.TextMessage, msg)
}

// notifyGameStart notifies both players involved in a game that the game has started.
func (wsm *WebSocketManager) notifyGameStart(game *models.Game) {
	if len(game.Players) != 2 {
		// Log error: Game must have exactly two players
		return
	}

	// Assign symbols and turns
	symbols := []string{"X", "O"} // First player is "X", second player is "O"
	for i, player := range game.Players {
		opponent := game.Players[1-i] // Get the other player as the opponent
		matchFoundPacket := models.MatchFoundPacket{
			BasePacket: models.BasePacket{Type: utils.GameStartPacketType},
			GameID:     game.ID,
			Opponent:   opponent.Username,
			YourSymbol: symbols[i],                     // Assign "X" to the first player and "O" to the second
			YourTurn:   game.CurrentTurn == symbols[i], // The first player ("X") starts the game
		}

		// Serialize the MatchFoundPacket to JSON
		msg, err := json.Marshal(matchFoundPacket)
		if err != nil {
			// Log error: Failed to marshal MatchFoundPacket
			continue
		}
		// Send the packet to the player's WebSocket connection
		if err := player.SendMessage(msg); err != nil {
			// Log error: Failed to send MatchFoundPacket
		}
	}
}

// sendNoMatchFound notifies a player that no match was found within the timeout.
func (wsm *WebSocketManager) sendNoMatchFound(conn *websocket.Conn) {
	// Construct and send a packet indicating no match was found
	noMatchPacket := models.BasePacket{
		Type: utils.NoMatchFoundType,
	}
	msg, err := json.Marshal(noMatchPacket)
	if err != nil {
		// Log error, handle failure to marshal packet
		return
	}
	conn.WriteMessage(websocket.TextMessage, msg)
}

// sendError sends an error message to the client.
func (wsm *WebSocketManager) sendError(conn *websocket.Conn, errorMsg string) {
	errorPacket := models.ErrorPacket{
		BasePacket: models.BasePacket{Type: utils.ErrorPacketType},
		Message:    errorMsg,
	}
	msg, _ := json.Marshal(errorPacket)
	conn.WriteMessage(websocket.TextMessage, msg)
}

// Add a new function to notify both players about the updated game state
func (wsm *WebSocketManager) notifyGameUpdate(game *models.Game) {
	for _, player := range game.Players {
		gameUpdatePacket := models.GameUpdatePacket{
			BasePacket:  models.BasePacket{Type: "gameUpdate"},
			GameID:      game.ID,
			Board:       game.Board,                   // Send the updated board
			CurrentTurn: game.IsPlayerCurrent(player), // Send the current turn
			Winner:      game.Winner,                  // Send the winner, if any
			Status:      game.Status,
		}
		msg, err := json.Marshal(gameUpdatePacket)
		if err != nil {
			// Log error, handle failure to marshal packet
			continue
		}
		player.SendMessage(msg)
	}
	if game.Status == utils.GameStateCompleted {
		// The game has ended, send "gameEnd" packets to both players
		var winner, loser *models.User
		if game.Winner == game.Players[0].Username {
			winner = game.Players[0]
			loser = game.Players[1]
		} else {
			winner = game.Players[1]
			loser = game.Players[0]
		}

		// Send "gameEnd" packet to the winner
		wsm.sendGameEndWinPacket(winner, game.ID, winner.Username)

		// Send "gameEnd" packet to the loser
		wsm.sendGameEndLosePacket(loser, game.ID, winner.Username)
	}
}

func (wsm *WebSocketManager) sendGameEndWinPacket(player *models.User, gameID string, winner string) {
	gameEndPacket := models.GameEndPacket{
		BasePacket: models.BasePacket{Type: "gameEnd"},
		GameID:     gameID,
		Winner:     winner,
		Outcome:    "win",
	}
	msg, err := json.Marshal(gameEndPacket)
	if err != nil {
		// Log error, handle failure to marshal packet
		return
	}
	player.SendMessage(msg)
}

func (wsm *WebSocketManager) sendGameEndLosePacket(player *models.User, gameID string, winner string) {
	gameEndPacket := models.GameEndPacket{
		BasePacket: models.BasePacket{Type: "gameEnd"},
		GameID:     gameID,
		Winner:     winner,
		Outcome:    "lose",
	}
	msg, err := json.Marshal(gameEndPacket)
	if err != nil {
		// Log error, handle failure to marshal packet
		return
	}
	player.SendMessage(msg)
}
