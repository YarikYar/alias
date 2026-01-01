package handlers

import (
	"encoding/json"
	"errors"
	"log"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/yaroslav/elias/internal/middleware"
	"github.com/yaroslav/elias/internal/models"
	"github.com/yaroslav/elias/internal/services"
	"github.com/yaroslav/elias/internal/ws"
)

type RoomHandler struct {
	roomService *services.RoomService
	gameService *services.GameService
	wordService *services.WordService
	hub         *ws.Hub
}

func NewRoomHandler(roomService *services.RoomService, gameService *services.GameService, wordService *services.WordService, hub *ws.Hub) *RoomHandler {
	return &RoomHandler{
		roomService: roomService,
		gameService: gameService,
		wordService: wordService,
		hub:         hub,
	}
}

func (h *RoomHandler) CreateRoom(c *fiber.Ctx) error {
	user := middleware.GetUser(c)
	if user == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "unauthorized"})
	}

	var req models.CreateRoomRequest
	if err := c.BodyParser(&req); err != nil {
		// If no body, use defaults
		req.Category = "general"
		req.NumTeams = 2
	}

	// Set defaults if not specified
	if req.Category == "" {
		req.Category = "general"
	}
	if req.NumTeams == 0 {
		req.NumTeams = 2
	}

	room, player, err := h.roomService.CreateRoom(c.Context(), user, req.Category, req.NumTeams)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"room":   room,
		"player": player,
	})
}

func (h *RoomHandler) GetRoom(c *fiber.Ctx) error {
	roomID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid room id"})
	}

	room, err := h.roomService.GetRoom(c.Context(), roomID)
	if err != nil {
		if errors.Is(err, services.ErrRoomNotFound) {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "room not found"})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	players, err := h.roomService.GetRoomPlayers(c.Context(), roomID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(models.RoomResponse{
		Room:    room,
		Players: players,
	})
}

func (h *RoomHandler) JoinRoom(c *fiber.Ctx) error {
	user := middleware.GetUser(c)
	if user == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "unauthorized"})
	}

	roomID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid room id"})
	}

	player, err := h.roomService.JoinRoom(c.Context(), roomID, user)
	if err != nil {
		if errors.Is(err, services.ErrRoomNotFound) {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "room not found"})
		}
		if errors.Is(err, services.ErrRoomFull) {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "room is full"})
		}
		if errors.Is(err, services.ErrGameInProgress) {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "game already in progress"})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	// Broadcast player joined to all clients in the room
	message := map[string]interface{}{
		"type": "player_joined",
		"payload": map[string]interface{}{
			"player": player,
		},
	}
	if msgBytes, err := json.Marshal(message); err == nil {
		h.hub.BroadcastToRoom(roomID, msgBytes)
	}

	return c.JSON(fiber.Map{"player": player})
}

func (h *RoomHandler) ChangeTeam(c *fiber.Ctx) error {
	user := middleware.GetUser(c)
	if user == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "unauthorized"})
	}

	roomID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid room id"})
	}

	var req models.ChangeTeamRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request"})
	}

	player, err := h.roomService.ChangeTeam(c.Context(), roomID, user.ID, req.Team)
	if err != nil {
		if errors.Is(err, services.ErrPlayerNotFound) {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "player not found"})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	// Broadcast team change to all clients in the room
	message := map[string]interface{}{
		"type": "team_changed",
		"payload": map[string]interface{}{
			"user_id": user.ID,
			"team":    req.Team,
		},
	}
	if msgBytes, err := json.Marshal(message); err == nil {
		h.hub.BroadcastToRoom(roomID, msgBytes)
	}

	return c.JSON(fiber.Map{"player": player})
}

func (h *RoomHandler) StartGame(c *fiber.Ctx) error {
	user := middleware.GetUser(c)
	if user == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "unauthorized"})
	}

	roomID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid room id"})
	}

	// Check if user is host
	isHost, err := h.roomService.IsHost(c.Context(), roomID, user.ID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	if !isHost {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "only host can start game"})
	}

	// Get players for the game
	players, err := h.roomService.GetRoomPlayers(c.Context(), roomID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	// Start game using GameService
	gameState, err := h.gameService.StartGame(c.Context(), roomID, players)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	// Get room to know category
	room, err := h.roomService.GetRoom(c.Context(), roomID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	// Get first word
	firstWord, err := h.wordService.GetRandomWord(c.Context(), roomID, "ru", room.Category)
	if err != nil {
		log.Printf("Failed to get random word: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to get first word"})
	}
	log.Printf("Got first word: %s (id=%d) category=%s", firstWord.Word, firstWord.ID, room.Category)

	// Set current word in game state
	if err := h.gameService.SetCurrentWord(c.Context(), roomID, firstWord); err != nil {
		log.Printf("Failed to set current word: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	// Broadcast game started to all clients in the room
	gameStartedMsg := map[string]interface{}{
		"type": "game_started",
		"payload": map[string]interface{}{
			"explainer_id": gameState.CurrentExplainer,
			"round_end_at": gameState.RoundEndAt.Unix(),
		},
	}
	if msgBytes, err := json.Marshal(gameStartedMsg); err == nil {
		log.Printf("Broadcasting game_started to room %s, explainer=%d", roomID, gameState.CurrentExplainer)
		h.hub.BroadcastToRoom(roomID, msgBytes)
	}

	// Broadcast first word to the room
	newWordMsg := map[string]interface{}{
		"type": "new_word",
		"payload": map[string]interface{}{
			"word_id": firstWord.ID,
			"word":    firstWord.Word,
		},
	}
	if msgBytes, err := json.Marshal(newWordMsg); err == nil {
		log.Printf("Broadcasting new_word to room %s: %s", roomID, firstWord.Word)
		h.hub.BroadcastToRoom(roomID, msgBytes)
	}

	// Start the round timer (60 seconds)
	log.Printf("Starting timer for room %s", roomID)
	h.hub.StartTimer(roomID, 60*time.Second)

	return c.JSON(fiber.Map{"status": "started"})
}

func (h *RoomHandler) GetStats(c *fiber.Ctx) error {
	roomID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid room id"})
	}

	room, err := h.roomService.GetRoom(c.Context(), roomID)
	if err != nil {
		if errors.Is(err, services.ErrRoomNotFound) {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "room not found"})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	players, err := h.roomService.GetRoomPlayers(c.Context(), roomID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	roundStats, err := h.wordService.GetRoundStats(c.Context(), roomID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	// Calculate team scores
	teamScores := map[string]int{"A": 0, "B": 0}
	var playerStats []*models.PlayerStats
	for _, p := range players {
		if p.Team != "" {
			teamScores[p.Team] += p.Score
		}
		playerStats = append(playerStats, &models.PlayerStats{
			UserID:    p.UserID,
			FirstName: p.FirstName,
			Team:      p.Team,
			Score:     p.Score,
		})
	}

	return c.JSON(models.GameStats{
		RoomID:     room.ID,
		TeamScores: teamScores,
		Players:    playerStats,
		Rounds:     roundStats,
	})
}
