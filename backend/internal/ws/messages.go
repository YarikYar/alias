package ws

import "github.com/yaroslav/elias/internal/models"

type MessageType string

const (
	// Client -> Server
	MsgTypeSwipe     MessageType = "swipe"
	MsgTypeVoteStart MessageType = "vote_start"
	MsgTypeVotePause MessageType = "vote_pause"

	// Server -> Client
	MsgTypePlayerJoined  MessageType = "player_joined"
	MsgTypePlayerLeft    MessageType = "player_left"
	MsgTypeTeamChanged   MessageType = "team_changed"
	MsgTypeGameStarted   MessageType = "game_started"
	MsgTypeNewWord       MessageType = "new_word"
	MsgTypeWordResult    MessageType = "word_result"
	MsgTypeTimer         MessageType = "timer"
	MsgTypeRoundEnd      MessageType = "round_end"
	MsgTypeGameEnd       MessageType = "game_end"
	MsgTypeError         MessageType = "error"
	MsgTypeRoomState     MessageType = "room_state"
	MsgTypeScoreUpdate   MessageType = "score_update"
)

type IncomingMessage struct {
	Type   MessageType `json:"type"`
	Action string      `json:"action,omitempty"`
}

type OutgoingMessage struct {
	Type    MessageType `json:"type"`
	Payload interface{} `json:"payload,omitempty"`
}

type PlayerJoinedPayload struct {
	Player *models.Player `json:"player"`
}

type PlayerLeftPayload struct {
	UserID int64 `json:"user_id"`
}

type TeamChangedPayload struct {
	UserID int64  `json:"user_id"`
	Team   string `json:"team"`
}

type GameStartedPayload struct {
	ExplainerID int64 `json:"explainer_id"`
	RoundEndAt  int64 `json:"round_end_at"`
}

type NewWordPayload struct {
	WordID int    `json:"word_id"`
	Word   string `json:"word"`
}

type WordResultPayload struct {
	WordID  int    `json:"word_id"`
	Word    string `json:"word"`
	Guessed bool   `json:"guessed"`
}

type TimerPayload struct {
	SecondsLeft int `json:"seconds_left"`
}

type RoundEndPayload struct {
	Round      int            `json:"round"`
	TeamScores map[string]int `json:"team_scores"`
	NextExplainer int64       `json:"next_explainer"`
}

type GameEndPayload struct {
	Winner     string         `json:"winner"`
	TeamScores map[string]int `json:"team_scores"`
}

type RoomStatePayload struct {
	Room    *models.Room     `json:"room"`
	Players []*models.Player `json:"players"`
}

type ScoreUpdatePayload struct {
	TeamScores map[string]int `json:"team_scores"`
}

type ErrorPayload struct {
	Message string `json:"message"`
}
