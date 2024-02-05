package managers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/websocket"
	"tictactoe/models"
)

// Helper function to create a WebSocket connection and simulate a client
func createWebSocketConnection(t *testing.T, wsm *WebSocketManager) (*websocket.Conn, *httptest.Server) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		wsm.HandleWebSocket(w, r)
	}))
	defer server.Close()

	// Connect to the WebSocket server
	url := "ws" + strings.TrimPrefix(server.URL, "http")
	conn, _, err := websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		t.Fatalf("Failed to connect to WebSocket server: %v", err)
	}

	return conn, server
}

func TestUserActions(t *testing.T) {
	// Create instances of UserManager, GameManager, MatchmakingManager, and WebSocketManager
	userManager := NewUserManager()
	gameManager := NewGameManager()
	matchmakingManager := NewMatchmakingManager()
	wsm := NewWebSocketManager(userManager, gameManager, matchmakingManager)

	// Create WebSocket connections for user1 and user2
	conn1, server1 := createWebSocketConnection(t, wsm)
	defer conn1.Close()
	defer server1.Close()

	conn2, server2 := createWebSocketConnection(t, wsm)
	defer conn2.Close()
	defer server2.Close()

	// Simulate user1's actions
	connectPacket1 := models.ConnectPacket{
		BasePacket: models.BasePacket{Type: "connect"},
		Username:   "user1",
		DeviceID:   "device1",
	}
	connectPacketJSON1, _ := json.Marshal(connectPacket1)
	conn1.WriteMessage(websocket.TextMessage, connectPacketJSON1)

	playPacket1 := models.PlayPacket{
		BasePacket: models.BasePacket{Type: "play"},
		Username:   "user1",
	}
	playPacketJSON1, _ := json.Marshal(playPacket1)
	conn1.WriteMessage(websocket.TextMessage, playPacketJSON1)

	movePacket1 := models.MovePacket{
		BasePacket: models.BasePacket{Type: "move"},
		GameID:     "gameID", // Provide a valid game ID
		Row:        0,
		Col:        0,
	}
	movePacketJSON1, _ := json.Marshal(movePacket1)
	conn1.WriteMessage(websocket.TextMessage, movePacketJSON1)

	// Simulate user2's actions
	connectPacket2 := models.ConnectPacket{
		BasePacket: models.BasePacket{Type: "connect"},
		Username:   "user2",
		DeviceID:   "device2",
	}
	connectPacketJSON2, _ := json.Marshal(connectPacket2)
	conn2.WriteMessage(websocket.TextMessage, connectPacketJSON2)

	playPacket2 := models.PlayPacket{
		BasePacket: models.BasePacket{Type: "play"},
		Username:   "user2",
	}
	playPacketJSON2, _ := json.Marshal(playPacket2)
	conn2.WriteMessage(websocket.TextMessage, playPacketJSON2)

	movePacket2 := models.MovePacket{
		BasePacket: models.BasePacket{Type: "move"},
		GameID:     "gameID", // Provide the same game ID as user1
		Row:        1,
		Col:        1,
	}
	movePacketJSON2, _ := json.Marshal(movePacket2)
	conn2.WriteMessage(websocket.TextMessage, movePacketJSON2)

	// Add assertions or checks to verify the expected behavior of the WebSocketManager
	// For example, check if the game state is updated correctly after user2's move

	// Wait for a brief moment to allow the WebSocketManager to process messages
	time.Sleep(100 * time.Millisecond)
}
