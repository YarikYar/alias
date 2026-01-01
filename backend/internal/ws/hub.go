package ws

import (
	"context"
	"encoding/json"
	"log"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"github.com/yaroslav/elias/internal/services"
)

type Hub struct {
	rooms       map[uuid.UUID]*RoomHub
	mu          sync.RWMutex
	rdb         *redis.Client
	gameService *services.GameService
	wordService *services.WordService
	roomService *services.RoomService
}

type RoomHub struct {
	roomID     uuid.UUID
	clients    map[int64]*Client
	broadcast  chan []byte
	register   chan *Client
	unregister chan *Client
	mu         sync.RWMutex
	hub        *Hub
	timer      *time.Ticker
	timerStop  chan struct{}
}

func NewHub(rdb *redis.Client, gameService *services.GameService, wordService *services.WordService, roomService *services.RoomService) *Hub {
	return &Hub{
		rooms:       make(map[uuid.UUID]*RoomHub),
		rdb:         rdb,
		gameService: gameService,
		wordService: wordService,
		roomService: roomService,
	}
}

func (h *Hub) Run() {
	// Subscribe to Redis pub/sub for cross-instance messaging
	ctx := context.Background()
	pubsub := h.rdb.Subscribe(ctx, "game_events")
	defer pubsub.Close()

	ch := pubsub.Channel()
	for msg := range ch {
		var event struct {
			RoomID  uuid.UUID       `json:"room_id"`
			Message json.RawMessage `json:"message"`
		}
		if err := json.Unmarshal([]byte(msg.Payload), &event); err != nil {
			continue
		}
		h.BroadcastToRoom(event.RoomID, event.Message)
	}
}

func (h *Hub) GetOrCreateRoomHub(roomID uuid.UUID) *RoomHub {
	h.mu.Lock()
	defer h.mu.Unlock()

	if room, ok := h.rooms[roomID]; ok {
		return room
	}

	room := &RoomHub{
		roomID:     roomID,
		clients:    make(map[int64]*Client),
		broadcast:  make(chan []byte, 256),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		hub:        h,
		timerStop:  make(chan struct{}),
	}
	h.rooms[roomID] = room
	go room.run()
	return room
}

func (h *Hub) Register(client *Client) {
	room := h.GetOrCreateRoomHub(client.roomID)
	room.register <- client
}

func (h *Hub) Unregister(client *Client) {
	h.mu.RLock()
	room, ok := h.rooms[client.roomID]
	h.mu.RUnlock()

	if ok {
		room.unregister <- client
	}
}

func (h *Hub) BroadcastToRoom(roomID uuid.UUID, message []byte) {
	h.mu.RLock()
	room, ok := h.rooms[roomID]
	h.mu.RUnlock()

	if ok {
		room.broadcast <- message
	}
}

func (h *Hub) SendToUser(roomID uuid.UUID, userID int64, message []byte) {
	h.mu.RLock()
	room, ok := h.rooms[roomID]
	h.mu.RUnlock()

	if ok {
		room.mu.RLock()
		client, ok := room.clients[userID]
		room.mu.RUnlock()

		if ok {
			select {
			case client.send <- message:
			default:
				close(client.send)
				room.mu.Lock()
				delete(room.clients, userID)
				room.mu.Unlock()
			}
		}
	}
}

func (h *Hub) StartTimer(roomID uuid.UUID, duration time.Duration) {
	h.mu.RLock()
	room, ok := h.rooms[roomID]
	h.mu.RUnlock()

	if ok {
		room.startTimer(duration)
	}
}

func (h *Hub) StopTimer(roomID uuid.UUID) {
	h.mu.RLock()
	room, ok := h.rooms[roomID]
	h.mu.RUnlock()

	if ok {
		room.stopTimer()
	}
}

func (rh *RoomHub) run() {
	for {
		select {
		case client := <-rh.register:
			rh.mu.Lock()
			rh.clients[client.user.ID] = client
			rh.mu.Unlock()
			log.Printf("Client %d joined room %s", client.user.ID, rh.roomID)

			// Send current game state to the newly connected client
			go rh.sendGameStateToClient(client)

		case client := <-rh.unregister:
			rh.mu.Lock()
			if _, ok := rh.clients[client.user.ID]; ok {
				delete(rh.clients, client.user.ID)
				close(client.send)
			}
			rh.mu.Unlock()
			log.Printf("Client %d left room %s", client.user.ID, rh.roomID)

			// Clean up empty rooms
			rh.mu.RLock()
			empty := len(rh.clients) == 0
			rh.mu.RUnlock()

			if empty {
				rh.stopTimer()
				rh.hub.mu.Lock()
				delete(rh.hub.rooms, rh.roomID)
				rh.hub.mu.Unlock()
				return
			}

		case message := <-rh.broadcast:
			rh.mu.RLock()
			for _, client := range rh.clients {
				select {
				case client.send <- message:
				default:
					close(client.send)
					delete(rh.clients, client.user.ID)
				}
			}
			rh.mu.RUnlock()
		}
	}
}

func (rh *RoomHub) startTimer(duration time.Duration) {
	rh.stopTimer()

	rh.timer = time.NewTicker(time.Second)
	endTime := time.Now().Add(duration)

	go func() {
		for {
			select {
			case <-rh.timerStop:
				return
			case t := <-rh.timer.C:
				remaining := int(endTime.Sub(t).Seconds())
				if remaining <= 0 {
					// Timer ended, handle round end
					rh.stopTimer()
					go rh.handleRoundEnd()
					return
				}

				msg, _ := json.Marshal(OutgoingMessage{
					Type:    MsgTypeTimer,
					Payload: TimerPayload{SecondsLeft: remaining},
				})
				rh.broadcast <- msg
			}
		}
	}()
}

func (rh *RoomHub) stopTimer() {
	if rh.timer != nil {
		rh.timer.Stop()
		rh.timer = nil
	}
	select {
	case rh.timerStop <- struct{}{}:
	default:
	}
}

func (rh *RoomHub) GetClientCount() int {
	rh.mu.RLock()
	defer rh.mu.RUnlock()
	return len(rh.clients)
}

func (rh *RoomHub) handleRoundEnd() {
	ctx := context.Background()

	// Get current game state
	gameState, err := rh.hub.gameService.GetGameState(ctx, rh.roomID)
	if err != nil || gameState == nil {
		log.Printf("Error getting game state: %v", err)
		return
	}

	// Check win condition
	hasWinner, winner, _ := rh.hub.gameService.CheckWinCondition(ctx, rh.roomID)

	if hasWinner {
		// Game ended - someone won
		if err := rh.hub.gameService.EndGame(ctx, rh.roomID); err != nil {
			log.Printf("Error ending game: %v", err)
			return
		}

		// Broadcast game end
		msg, _ := json.Marshal(OutgoingMessage{
			Type: MsgTypeGameEnd,
			Payload: GameEndPayload{
				Winner:     winner,
				TeamScores: gameState.TeamScores,
			},
		})
		rh.broadcast <- msg
		log.Printf("Game ended in room %s, winner: %s", rh.roomID, winner)
	} else {
		// Get room players for next round
		players, err := rh.hub.roomService.GetRoomPlayers(ctx, rh.roomID)
		if err != nil {
			log.Printf("Error getting room players: %v", err)
			return
		}

		// Start next round
		nextState, err := rh.hub.gameService.NextRound(ctx, rh.roomID, players)
		if err != nil {
			log.Printf("Error starting next round: %v", err)
			return
		}

		// Broadcast round end
		msg, _ := json.Marshal(OutgoingMessage{
			Type: MsgTypeRoundEnd,
			Payload: RoundEndPayload{
				Round:         nextState.CurrentRound - 1,
				TeamScores:    nextState.TeamScores,
				NextExplainer: nextState.CurrentExplainer,
			},
		})
		rh.broadcast <- msg

		// Get room category
		room, err := rh.hub.roomService.GetRoom(ctx, rh.roomID)
		if err != nil {
			log.Printf("Error getting room: %v", err)
			return
		}

		// Get first word for next round
		nextWord, err := rh.hub.wordService.GetRandomWord(ctx, rh.roomID, "ru", room.Category)
		if err != nil {
			log.Printf("Error getting next word: %v", err)
			return
		}

		// Set current word
		if err := rh.hub.gameService.SetCurrentWord(ctx, rh.roomID, nextWord); err != nil {
			log.Printf("Error setting current word: %v", err)
			return
		}

		// Broadcast new word
		newWordMsg, _ := json.Marshal(OutgoingMessage{
			Type: MsgTypeNewWord,
			Payload: NewWordPayload{
				WordID: nextWord.ID,
				Word:   nextWord.Word,
			},
		})
		rh.broadcast <- newWordMsg

		// Start timer for next round
		rh.startTimer(60 * time.Second)
		log.Printf("Started round %d in room %s, explainer: %d", nextState.CurrentRound, rh.roomID, nextState.CurrentExplainer)
	}
}

func (rh *RoomHub) sendGameStateToClient(client *Client) {
	ctx := context.Background()

	// Get current game state
	gameState, err := rh.hub.gameService.GetGameState(ctx, rh.roomID)
	if err != nil || gameState == nil {
		return
	}

	// Only send state if game is playing
	if gameState.Status != "playing" {
		return
	}

	// Send game_started event with current explainer
	gameStartedMsg, _ := json.Marshal(OutgoingMessage{
		Type: MsgTypeGameStarted,
		Payload: GameStartedPayload{
			ExplainerID: gameState.CurrentExplainer,
			RoundEndAt:  gameState.RoundEndAt.Unix(),
		},
	})
	client.send <- gameStartedMsg

	// Send current word if exists
	if gameState.CurrentWord != nil {
		newWordMsg, _ := json.Marshal(OutgoingMessage{
			Type: MsgTypeNewWord,
			Payload: NewWordPayload{
				WordID: gameState.CurrentWord.ID,
				Word:   gameState.CurrentWord.Word,
			},
		})
		client.send <- newWordMsg
	}

	// Send current scores
	scoreMsg, _ := json.Marshal(OutgoingMessage{
		Type: MsgTypeScoreUpdate,
		Payload: ScoreUpdatePayload{
			TeamScores: gameState.TeamScores,
		},
	})
	client.send <- scoreMsg

	// Restart timer if not running
	remaining := gameState.RoundEndAt.Sub(time.Now())
	if remaining > 0 {
		// Check if timer is running, if not - start it
		if rh.timer == nil {
			rh.startTimer(remaining)
		}

		// Send current timer value
		timerMsg, _ := json.Marshal(OutgoingMessage{
			Type: MsgTypeTimer,
			Payload: TimerPayload{
				SecondsLeft: int(remaining.Seconds()),
			},
		})
		client.send <- timerMsg
	}
}
