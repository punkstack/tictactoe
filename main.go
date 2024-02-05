package main

import (
	"log"
	"net/http"
	"tictactoe/managers"
)

func main() {
	// Initialize managers
	userManager := managers.NewUserManager()
	gameManager := managers.NewGameManager()
	matchmakingManager := managers.NewMatchmakingManager()

	// Initialize WebSocketManager with references to other managers
	websocketManager := managers.NewWebSocketManager(userManager, gameManager, matchmakingManager)

	// Setup WebSocket handler
	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		websocketManager.HandleWebSocket(w, r)
	})

	// Start listening for WebSocket connections
	log.Println("WebSocket server starting on :8080") // Log before starting the server
	err := http.ListenAndServe(":8080", nil)          // Start the server
	if err != nil {
		log.Fatalf("Failed to start WebSocket server: %s", err) // This will log only if there's an error starting the server
	}
}
