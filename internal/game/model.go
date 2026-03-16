package game

// Game represents a rock-paper-scissors game.
type Game struct {
	ID          string  `json:"id"`
	Player1Move *string `json:"player1_move"`
	Player2Move *string `json:"player2_move"`
	Status      string  `json:"status"`
	Winner      *string `json:"winner"`
	CreatedAt   string  `json:"created_at"`
}
