package services

import (
	"context"
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
	"github.com/yaroslav/elias/internal/models"
)

const (
	RoundDuration    = 60 * time.Second
	MaxWordsPerRound = 5
	WinningScore     = 20
)

type GameState struct {
	RoomID           uuid.UUID `json:"room_id"`
	Status           string    `json:"status"`
	CurrentRound     int       `json:"current_round"`
	CurrentExplainer int64     `json:"current_explainer"`
	CurrentWord      *WordState `json:"current_word,omitempty"`
	RoundEndAt       time.Time `json:"round_end_at"`
	WordsThisRound   int       `json:"words_this_round"`
	TeamAScore       int       `json:"team_a_score"`
	TeamBScore       int       `json:"team_b_score"`
}

type WordState struct {
	ID   int    `json:"id"`
	Word string `json:"word"`
}

type GameService struct {
	pool *pgxpool.Pool
	rdb  *redis.Client
}

func NewGameService(pool *pgxpool.Pool, rdb *redis.Client) *GameService {
	return &GameService{pool: pool, rdb: rdb}
}

func (s *GameService) GetGameState(ctx context.Context, roomID uuid.UUID) (*GameState, error) {
	key := "game:" + roomID.String()
	data, err := s.rdb.Get(ctx, key).Bytes()
	if err != nil {
		if err == redis.Nil {
			return nil, nil
		}
		return nil, err
	}

	var state GameState
	if err := json.Unmarshal(data, &state); err != nil {
		return nil, err
	}
	return &state, nil
}

func (s *GameService) SaveGameState(ctx context.Context, state *GameState) error {
	key := "game:" + state.RoomID.String()
	data, err := json.Marshal(state)
	if err != nil {
		return err
	}
	return s.rdb.Set(ctx, key, data, 24*time.Hour).Err()
}

func (s *GameService) StartGame(ctx context.Context, roomID uuid.UUID, players []*models.Player) (*GameState, error) {
	// Find first explainer (first player from team A or first player)
	var firstExplainer int64
	for _, p := range players {
		if p.Team == "A" {
			firstExplainer = p.UserID
			break
		}
	}
	if firstExplainer == 0 && len(players) > 0 {
		firstExplainer = players[0].UserID
	}

	// Calculate initial scores
	teamAScore := 0
	teamBScore := 0
	for _, p := range players {
		if p.Team == "A" {
			teamAScore += p.Score
		} else if p.Team == "B" {
			teamBScore += p.Score
		}
	}

	state := &GameState{
		RoomID:           roomID,
		Status:           string(models.RoomStatusPlaying),
		CurrentRound:     1,
		CurrentExplainer: firstExplainer,
		RoundEndAt:       time.Now().Add(RoundDuration),
		WordsThisRound:   0,
		TeamAScore:       teamAScore,
		TeamBScore:       teamBScore,
	}

	if err := s.SaveGameState(ctx, state); err != nil {
		return nil, err
	}

	// Update room status in DB
	_, err := s.pool.Exec(ctx, `
		UPDATE rooms
		SET status = $1, current_round = $2, current_explainer_id = $3, round_end_at = $4
		WHERE id = $5
	`, models.RoomStatusPlaying, state.CurrentRound, state.CurrentExplainer, state.RoundEndAt, roomID)
	if err != nil {
		return nil, err
	}

	return state, nil
}

func (s *GameService) SetCurrentWord(ctx context.Context, roomID uuid.UUID, word *models.Word) error {
	state, err := s.GetGameState(ctx, roomID)
	if err != nil {
		return err
	}
	if state == nil {
		return ErrRoomNotFound
	}

	state.CurrentWord = &WordState{
		ID:   word.ID,
		Word: word.Word,
	}
	return s.SaveGameState(ctx, state)
}

func (s *GameService) ProcessSwipe(ctx context.Context, roomID uuid.UUID, userID int64, action string) (bool, *models.Word, error) {
	state, err := s.GetGameState(ctx, roomID)
	if err != nil {
		return false, nil, err
	}
	if state == nil {
		return false, nil, ErrRoomNotFound
	}

	// Only explainer can swipe
	if state.CurrentExplainer != userID {
		return false, nil, nil
	}

	if state.CurrentWord == nil {
		return false, nil, nil
	}

	guessed := action == "up"
	word := state.CurrentWord

	// Record result
	_, err = s.pool.Exec(ctx, `
		INSERT INTO round_words (room_id, word_id, round_num, guessed)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT DO NOTHING
	`, roomID, word.ID, state.CurrentRound, guessed)
	if err != nil {
		return false, nil, err
	}

	// Update player score if guessed
	if guessed {
		_, err = s.pool.Exec(ctx, `
			UPDATE players SET score = score + 1
			WHERE room_id = $1 AND user_id = $2
		`, roomID, userID)
		if err != nil {
			return false, nil, err
		}

		// Update team score
		var team string
		err = s.pool.QueryRow(ctx, `
			SELECT team FROM players WHERE room_id = $1 AND user_id = $2
		`, roomID, userID).Scan(&team)
		if err == nil {
			if team == "A" {
				state.TeamAScore++
			} else if team == "B" {
				state.TeamBScore++
			}
		}
	}

	state.WordsThisRound++
	state.CurrentWord = nil

	if err := s.SaveGameState(ctx, state); err != nil {
		return false, nil, err
	}

	return guessed, &models.Word{ID: word.ID, Word: word.Word}, nil
}

func (s *GameService) NextRound(ctx context.Context, roomID uuid.UUID, players []*models.Player) (*GameState, error) {
	state, err := s.GetGameState(ctx, roomID)
	if err != nil {
		return nil, err
	}
	if state == nil {
		return nil, ErrRoomNotFound
	}

	// Find next explainer (rotate through players)
	var nextExplainer int64
	foundCurrent := false
	for _, p := range players {
		if foundCurrent && p.Team != "" {
			nextExplainer = p.UserID
			break
		}
		if p.UserID == state.CurrentExplainer {
			foundCurrent = true
		}
	}
	// Wrap around if needed
	if nextExplainer == 0 {
		for _, p := range players {
			if p.Team != "" {
				nextExplainer = p.UserID
				break
			}
		}
	}

	state.CurrentRound++
	state.CurrentExplainer = nextExplainer
	state.RoundEndAt = time.Now().Add(RoundDuration)
	state.WordsThisRound = 0
	state.CurrentWord = nil

	if err := s.SaveGameState(ctx, state); err != nil {
		return nil, err
	}

	// Update DB
	_, err = s.pool.Exec(ctx, `
		UPDATE rooms
		SET current_round = $1, current_explainer_id = $2, round_end_at = $3
		WHERE id = $4
	`, state.CurrentRound, state.CurrentExplainer, state.RoundEndAt, roomID)

	return state, err
}

func (s *GameService) EndGame(ctx context.Context, roomID uuid.UUID) error {
	state, err := s.GetGameState(ctx, roomID)
	if err != nil {
		return err
	}
	if state != nil {
		state.Status = string(models.RoomStatusFinished)
		if err := s.SaveGameState(ctx, state); err != nil {
			return err
		}
	}

	_, err = s.pool.Exec(ctx, `
		UPDATE rooms SET status = $1 WHERE id = $2
	`, models.RoomStatusFinished, roomID)
	return err
}

func (s *GameService) CheckWinCondition(ctx context.Context, roomID uuid.UUID) (bool, string, error) {
	state, err := s.GetGameState(ctx, roomID)
	if err != nil {
		return false, "", err
	}
	if state == nil {
		return false, "", nil
	}

	if state.TeamAScore >= WinningScore {
		return true, "A", nil
	}
	if state.TeamBScore >= WinningScore {
		return true, "B", nil
	}
	return false, "", nil
}

func (s *GameService) GetTeamScores(ctx context.Context, roomID uuid.UUID) (int, int, error) {
	state, err := s.GetGameState(ctx, roomID)
	if err != nil {
		return 0, 0, err
	}
	if state == nil {
		return 0, 0, nil
	}
	return state.TeamAScore, state.TeamBScore, nil
}
