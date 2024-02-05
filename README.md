# Tic-Tac-Toe WebSocket Game Server

This is a WebSocket-based Tic-Tac-Toe game server implemented in Go. It allows multiple clients to connect, play games, and enjoy real-time updates.

## Folder Structure
```
tictactoe/
├── managers/
│ ├── user_manager.go
│ ├── game_manager.go
│ ├── matchmaking_manager.go
│ └── websocket_manager.go
├── models/
│ ├── user.go
│ ├── game.go
│ └── packets.go
├── utils/
│ ├── constants.go
│ └── helpers.go
├── main.go
├── Makefile
└── README.md
```
- **managers**: Contains the main logic for managing users, games, matchmaking, and WebSocket connections.
- **models**: Defines the data models for users, games, and packets sent over WebSocket.
- **utils**: Contains utility functions and constants used throughout the project.

## Getting Started

1. Clone the repository:

```bash
git clone https://github.com/yourusername/tictactoe-websocket.git
cd tictactoe-websocket
make build
make run
```

The WebSocket server will start and listen on port 8080.

## Makefile Commands

- `make build`: Build the project.
- `make run`: Start the WebSocket server.
- `make test`: Run tests for the project.
- `make debug`: Run the server in debug mode.

## Managers

- **User Manager**: Manages user registration, disconnection, and statistics tracking.
- **Game Manager**: Handles game creation, updates, and game state.
- **Matchmaking Manager**: Implements game matchmaking logic.
- **WebSocket Manager**: Manages WebSocket connections, message handling, and game notifications.

For more details, refer to the source code in the `managers` directory.

## Contributing

Feel free to contribute to this project by submitting issues or pull requests. We welcome any improvements, bug fixes, or new features.

Happy gaming!





