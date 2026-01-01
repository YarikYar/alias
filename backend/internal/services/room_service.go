package services

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/yaroslav/elias/internal/models"
)

var (
	ErrRoomNotFound   = errors.New("room not found")
	ErrPlayerNotFound = errors.New("player not found")
	ErrAlreadyInRoom  = errors.New("player already in room")
	ErrRoomFull       = errors.New("room is full")
	ErrNotHost        = errors.New("only host can perform this action")
	ErrGameInProgress = errors.New("game already in progress")
)

type RoomService struct {
	pool *pgxpool.Pool
}

func NewRoomService(pool *pgxpool.Pool) *RoomService {
	return &RoomService{pool: pool}
}

func (s *RoomService) CreateRoom(ctx context.Context, user *models.TelegramUser, category string, numTeams int) (*models.Room, *models.Player, error) {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return nil, nil, err
	}
	defer tx.Rollback(ctx)

	// Default category if not specified
	if category == "" {
		category = "general"
	}

	// Default num_teams if not specified or invalid
	if numTeams < 2 {
		numTeams = 2
	}
	if numTeams > 5 {
		numTeams = 5
	}

	// Create room
	room := models.Room{
		CurrentRound:       0,
		Status:             models.RoomStatusLobby,
		CurrentExplainerID: nil,
		RoundEndAt:         nil,
		Category:           category,
		NumTeams:           numTeams,
	}
	err = tx.QueryRow(ctx, `
		INSERT INTO rooms (status, current_round, category, num_teams) VALUES ($1, $2, $3, $4)
		RETURNING id, created_at
	`, models.RoomStatusLobby, 0, category, numTeams).Scan(&room.ID, &room.CreatedAt)
	if err != nil {
		return nil, nil, err
	}

	// Add creator as host
	player := models.Player{
		RoomID:    room.ID,
		UserID:    user.ID,
		Username:  user.Username,
		FirstName: user.FirstName,
		Team:      "",
		Score:     0,
		IsHost:    true,
	}
	err = tx.QueryRow(ctx, `
		INSERT INTO players (room_id, user_id, username, first_name, is_host)
		VALUES ($1, $2, NULLIF($3, ''), NULLIF($4, ''), TRUE)
		RETURNING id, joined_at
	`, room.ID, user.ID, user.Username, user.FirstName).Scan(&player.ID, &player.JoinedAt)
	if err != nil {
		return nil, nil, err
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, nil, err
	}

	return &room, &player, nil
}

func (s *RoomService) GetRoom(ctx context.Context, roomID uuid.UUID) (*models.Room, error) {
	room := &models.Room{}
	err := s.pool.QueryRow(ctx, `
		SELECT id, status, current_round, category, num_teams, created_at
		FROM rooms WHERE id = $1
	`, roomID).Scan(&room.ID, &room.Status, &room.CurrentRound, &room.Category, &room.NumTeams, &room.CreatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrRoomNotFound
		}
		return nil, err
	}
	return room, nil
}

func (s *RoomService) GetRoomPlayers(ctx context.Context, roomID uuid.UUID) ([]*models.Player, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT id, room_id, user_id, COALESCE(username, ''), COALESCE(first_name, ''), COALESCE(team, ''), score, is_host, joined_at
		FROM players WHERE room_id = $1
		ORDER BY joined_at
	`, roomID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var players []*models.Player
	for rows.Next() {
		var p models.Player
		if err := rows.Scan(
			&p.ID, &p.RoomID, &p.UserID, &p.Username,
			&p.FirstName, &p.Team, &p.Score, &p.IsHost, &p.JoinedAt,
		); err != nil {
			return nil, err
		}
		players = append(players, &p)
	}
	return players, nil
}

func (s *RoomService) JoinRoom(ctx context.Context, roomID uuid.UUID, user *models.TelegramUser) (*models.Player, error) {
	// Check room exists and is in lobby
	room, err := s.GetRoom(ctx, roomID)
	if err != nil {
		return nil, err
	}
	if room.Status != models.RoomStatusLobby {
		return nil, ErrGameInProgress
	}

	// Check player count
	players, err := s.GetRoomPlayers(ctx, roomID)
	if err != nil {
		return nil, err
	}
	if len(players) >= 8 {
		return nil, ErrRoomFull
	}

	// Check if already in room
	for _, p := range players {
		if p.UserID == user.ID {
			return p, nil // Return existing player
		}
	}

	// Add player
	player := models.Player{
		RoomID:    roomID,
		UserID:    user.ID,
		Username:  user.Username,
		FirstName: user.FirstName,
		Team:      "",
		Score:     0,
		IsHost:    false,
	}
	err = s.pool.QueryRow(ctx, `
		INSERT INTO players (room_id, user_id, username, first_name)
		VALUES ($1, $2, NULLIF($3, ''), NULLIF($4, ''))
		RETURNING id, joined_at
	`, roomID, user.ID, user.Username, user.FirstName).Scan(&player.ID, &player.JoinedAt)
	if err != nil {
		return nil, err
	}

	return &player, nil
}

func (s *RoomService) ChangeTeam(ctx context.Context, roomID uuid.UUID, userID int64, team string) (*models.Player, error) {
	if team != "A" && team != "B" && team != "" {
		return nil, errors.New("invalid team")
	}

	var player models.Player
	err := s.pool.QueryRow(ctx, `
		UPDATE players SET team = $1
		WHERE room_id = $2 AND user_id = $3
		RETURNING id, room_id, user_id, username, first_name, team, score, is_host, joined_at
	`, team, roomID, userID).Scan(
		&player.ID, &player.RoomID, &player.UserID, &player.Username,
		&player.FirstName, &player.Team, &player.Score, &player.IsHost, &player.JoinedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrPlayerNotFound
		}
		return nil, err
	}

	return &player, nil
}

func (s *RoomService) GetPlayer(ctx context.Context, roomID uuid.UUID, userID int64) (*models.Player, error) {
	var player models.Player
	err := s.pool.QueryRow(ctx, `
		SELECT id, room_id, user_id, username, first_name, team, score, is_host, joined_at
		FROM players WHERE room_id = $1 AND user_id = $2
	`, roomID, userID).Scan(
		&player.ID, &player.RoomID, &player.UserID, &player.Username,
		&player.FirstName, &player.Team, &player.Score, &player.IsHost, &player.JoinedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrPlayerNotFound
		}
		return nil, err
	}
	return &player, nil
}

func (s *RoomService) IsHost(ctx context.Context, roomID uuid.UUID, userID int64) (bool, error) {
	player, err := s.GetPlayer(ctx, roomID, userID)
	if err != nil {
		return false, err
	}
	return player.IsHost, nil
}

func (s *RoomService) UpdateRoomStatus(ctx context.Context, roomID uuid.UUID, status models.RoomStatus) error {
	_, err := s.pool.Exec(ctx, `
		UPDATE rooms SET status = $1 WHERE id = $2
	`, status, roomID)
	return err
}

func (s *RoomService) UpdateScore(ctx context.Context, roomID uuid.UUID, userID int64, delta int) error {
	_, err := s.pool.Exec(ctx, `
		UPDATE players SET score = score + $1
		WHERE room_id = $2 AND user_id = $3
	`, delta, roomID, userID)
	return err
}
