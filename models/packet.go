package models

// BasePacket defines the basic structure of all packets with a common Type field.
type BasePacket struct {
	Type string `json:"type"`
}

// ConnectPacket is sent by the client to establish a user session.
type ConnectPacket struct {
	BasePacket
	Username string `json:"username"`
	DeviceID string `json:"deviceId"`
}

// PlayPacket is sent by the client to request starting or joining a game.
type PlayPacket struct {
	BasePacket
	Username string `json:"username"`
}

// MovePacket is sent by the client when making a move in a game.
type MovePacket struct {
	BasePacket
	GameID string `json:"gameId"`
	Row    int    `json:"row"`
	Col    int    `json:"col"`
}

// GameUpdatePacket is sent by the server to inform clients about the current game state.
type GameUpdatePacket struct {
	BasePacket
	GameID      string       `json:"gameId"`
	Board       [3][3]string `json:"board"`
	CurrentTurn bool         `json:"currentTurn"`
	Winner      string       `json:"winner,omitempty"` // Empty if the game is ongoing
	Status      string       `json:"status"`
}

// MatchFoundPacket is sent by the server to notify the client that a match has been found.
type MatchFoundPacket struct {
	BasePacket
	GameID     string `json:"gameId"`
	Opponent   string `json:"opponent"`
	YourSymbol string `json:"yourSymbol"` // "X" or "O"
	YourTurn   bool   `json:"yourTurn"`
}

// ErrorPacket is sent by the server in response to errors.
type ErrorPacket struct {
	BasePacket
	Message string `json:"message"`
}

// GameEndPacket is sent by the server in response to game end.
type GameEndPacket struct {
	BasePacket
	GameID  string `json:"gameId"`
	Winner  string `json:"winner"`
	Outcome string `json:"outcome"` // Possible values: "win", "lose", "draw"
}
