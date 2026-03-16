package game

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

var (
	ErrInvalidArgument = errors.New("invalid argument")
	ErrNotFound        = errors.New("not found")
)

var validMoves = []string{"rock", "paper", "scissors"}

// beats[attacker][defender] = true means attacker wins
var beats = map[string]map[string]bool{
	"rock":     {"scissors": true},
	"scissors": {"paper": true},
	"paper":    {"rock": true},
}

// Service manages game creation, moves, and persistence.
type Service struct {
	mu          sync.Mutex
	storageFile string
}

// NewService creates a new game service with the given data directory.
func NewService(dataDir string) (*Service, error) {
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		return nil, fmt.Errorf("create data dir: %w", err)
	}
	return &Service{
		storageFile: filepath.Join(dataDir, "games.json"),
	}, nil
}

// CreateGame creates a new game and persists it.
func (s *Service) CreateGame() (*Game, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	b := make([]byte, 8)
	if _, err := rand.Read(b); err != nil {
		return nil, fmt.Errorf("generate id: %w", err)
	}

	now := time.Now()
	// PHP DateTimeInterface::ATOM uses +00:00 offset format, not Z
	createdAt := now.Format("2006-01-02T15:04:05-07:00")

	game := &Game{
		ID:        hex.EncodeToString(b),
		Status:    "waiting",
		CreatedAt: createdAt,
	}

	games, err := s.loadGames()
	if err != nil {
		return nil, err
	}

	games[game.ID] = game
	if err := s.saveGames(games); err != nil {
		return nil, err
	}

	return game, nil
}

// MakeMove applies a player's move to a game.
func (s *Service) MakeMove(gameID string, player int, choice string) (*Game, error) {
	if player != 1 && player != 2 {
		return nil, fmt.Errorf("%w: Player must be 1 or 2.", ErrInvalidArgument)
	}

	validChoice := false
	for _, m := range validMoves {
		if choice == m {
			validChoice = true
			break
		}
	}
	if !validChoice {
		return nil, fmt.Errorf("%w: Choice must be one of: rock, paper, scissors.", ErrInvalidArgument)
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	games, err := s.loadGames()
	if err != nil {
		return nil, err
	}

	game, ok := games[gameID]
	if !ok {
		return nil, fmt.Errorf("%w: Game not found.", ErrNotFound)
	}

	if game.Status == "complete" {
		return nil, fmt.Errorf("%w: Game is already complete.", ErrInvalidArgument)
	}

	if player == 1 {
		if game.Player1Move != nil {
			return nil, fmt.Errorf("%w: Player 1 has already moved.", ErrInvalidArgument)
		}
		game.Player1Move = &choice
	} else {
		if game.Player2Move != nil {
			return nil, fmt.Errorf("%w: Player 2 has already moved.", ErrInvalidArgument)
		}
		game.Player2Move = &choice
	}

	if game.Player1Move != nil && game.Player2Move != nil {
		game.Status = "complete"
		winner := determineWinner(*game.Player1Move, *game.Player2Move)
		game.Winner = &winner
	}

	games[gameID] = game
	if err := s.saveGames(games); err != nil {
		return nil, err
	}

	return game, nil
}

func determineWinner(p1, p2 string) string {
	if p1 == p2 {
		return "draw"
	}
	if beats[p1][p2] {
		return "player1"
	}
	return "player2"
}

func (s *Service) loadGames() (map[string]*Game, error) {
	data, err := os.ReadFile(s.storageFile)
	if err != nil {
		if os.IsNotExist(err) {
			return make(map[string]*Game), nil
		}
		return nil, fmt.Errorf("read games: %w", err)
	}
	if len(data) == 0 {
		return make(map[string]*Game), nil
	}

	var games map[string]*Game
	if err := json.Unmarshal(data, &games); err != nil {
		return make(map[string]*Game), nil
	}
	if games == nil {
		games = make(map[string]*Game)
	}
	return games, nil
}

func (s *Service) saveGames(games map[string]*Game) error {
	data, err := json.MarshalIndent(games, "", "    ")
	if err != nil {
		return fmt.Errorf("marshal games: %w", err)
	}
	if err := os.WriteFile(s.storageFile, data, 0644); err != nil {
		return fmt.Errorf("write games: %w", err)
	}
	return nil
}
