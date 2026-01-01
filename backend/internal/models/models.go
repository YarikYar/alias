package models

import (
	"time"

	"github.com/google/uuid"
)

type RoomStatus string

const (
	RoomStatusLobby    RoomStatus = "lobby"
	RoomStatusPlaying  RoomStatus = "playing"
	RoomStatusFinished RoomStatus = "finished"
)

type Room struct {
	ID                 uuid.UUID  `json:"id"`
	Status             RoomStatus `json:"status"`
	CurrentRound       int        `json:"current_round"`
	CurrentExplainerID *int64     `json:"current_explainer_id,omitempty"`
	RoundEndAt         *time.Time `json:"round_end_at,omitempty"`
	Category           string     `json:"category"`
	NumTeams           int        `json:"num_teams"`
	CreatedAt          time.Time  `json:"created_at"`
}

type Player struct {
	ID        int       `json:"id"`
	RoomID    uuid.UUID `json:"room_id"`
	UserID    int64     `json:"user_id"`
	Username  string    `json:"username,omitempty"`
	FirstName string    `json:"first_name,omitempty"`
	Team      string    `json:"team,omitempty"`
	Score     int       `json:"score"`
	IsHost    bool      `json:"is_host"`
	JoinedAt  time.Time `json:"joined_at"`
}

type Word struct {
	ID       int    `json:"id"`
	Word     string `json:"word"`
	Lang     string `json:"lang"`
	Category string `json:"category"`
}

type RoundWord struct {
	ID        int       `json:"id"`
	RoomID    uuid.UUID `json:"room_id"`
	WordID    int       `json:"word_id"`
	RoundNum  int       `json:"round_num"`
	Guessed   bool      `json:"guessed"`
	CreatedAt time.Time `json:"created_at"`
}

// Telegram user from initData
type TelegramUser struct {
	ID        int64  `json:"id"`
	Username  string `json:"username,omitempty"`
	FirstName string `json:"first_name,omitempty"`
	LastName  string `json:"last_name,omitempty"`
	PhotoURL  string `json:"photo_url,omitempty"`
}

// API responses
type RoomResponse struct {
	Room    *Room     `json:"room"`
	Players []*Player `json:"players"`
}

type CreateRoomRequest struct {
	Category string `json:"category"`
	NumTeams int    `json:"num_teams"`
}

type JoinRoomRequest struct{}

type ChangeTeamRequest struct {
	Team string `json:"team"`
}

type GameStats struct {
	RoomID     uuid.UUID            `json:"room_id"`
	TeamScores map[string]int       `json:"team_scores"`
	Players    []*PlayerStats       `json:"players"`
	Rounds     []*RoundStats        `json:"rounds"`
}

type PlayerStats struct {
	UserID      int64  `json:"user_id"`
	FirstName   string `json:"first_name"`
	Team        string `json:"team"`
	Score       int    `json:"score"`
	WordsGuessed int   `json:"words_guessed"`
	WordsMissed  int   `json:"words_missed"`
}

type RoundStats struct {
	RoundNum     int `json:"round_num"`
	ExplainerID  int64 `json:"explainer_id"`
	WordsGuessed int   `json:"words_guessed"`
	WordsMissed  int   `json:"words_missed"`
}
