package ws

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"github.com/gofiber/contrib/websocket"
	"github.com/google/uuid"
	"github.com/yaroslav/elias/internal/models"
)

const (
	writeWait      = 10 * time.Second
	pongWait       = 60 * time.Second
	pingPeriod     = (pongWait * 9) / 10
	maxMessageSize = 512
)

type Client struct {
	hub    *Hub
	conn   *websocket.Conn
	send   chan []byte
	roomID uuid.UUID
	user   *models.TelegramUser
}

func NewClient(hub *Hub, conn *websocket.Conn, roomID uuid.UUID, user *models.TelegramUser) *Client {
	return &Client{
		hub:    hub,
		conn:   conn,
		send:   make(chan []byte, 256),
		roomID: roomID,
		user:   user,
	}
}

func (c *Client) ReadPump() {
	defer func() {
		c.hub.Unregister(c)
		c.conn.Close()
	}()

	c.conn.SetReadLimit(maxMessageSize)
	c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error {
		c.conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("error: %v", err)
			}
			break
		}

		var msg IncomingMessage
		if err := json.Unmarshal(message, &msg); err != nil {
			continue
		}

		c.handleMessage(&msg)
	}
}

func (c *Client) WritePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			if err := c.conn.WriteMessage(websocket.TextMessage, message); err != nil {
				return
			}

		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

func (c *Client) handleMessage(msg *IncomingMessage) {
	switch msg.Type {
	case MsgTypeSwipe:
		c.handleSwipe(msg.Action)
	case MsgTypeVoteStart:
		c.handleVoteStart()
	case MsgTypeVotePause:
		c.handleVotePause()
	}
}

func (c *Client) handleSwipe(action string) {
	log.Printf("Player %d swiped %s in room %s", c.user.ID, action, c.roomID)

	// Only process up/down swipes
	if action != "up" && action != "down" {
		return
	}

	// Process swipe through game service
	ctx := context.Background()
	guessed, word, err := c.hub.gameService.ProcessSwipe(ctx, c.roomID, c.user.ID, action)
	if err != nil {
		log.Printf("Error processing swipe: %v", err)
		return
	}

	// If word was processed, broadcast result
	if word != nil {
		// Broadcast word result
		resultMsg, _ := json.Marshal(OutgoingMessage{
			Type: MsgTypeWordResult,
			Payload: WordResultPayload{
				WordID:  word.ID,
				Word:    word.Word,
				Guessed: guessed,
			},
		})
		c.hub.BroadcastToRoom(c.roomID, resultMsg)

		// Get and broadcast team scores
		scoreA, scoreB, _ := c.hub.gameService.GetTeamScores(ctx, c.roomID)
		scoreMsg, _ := json.Marshal(OutgoingMessage{
			Type: MsgTypeScoreUpdate,
			Payload: ScoreUpdatePayload{
				TeamScores: map[string]int{"A": scoreA, "B": scoreB},
			},
		})
		c.hub.BroadcastToRoom(c.roomID, scoreMsg)

		// Get room category
		room, err := c.hub.roomService.GetRoom(ctx, c.roomID)
		if err != nil {
			log.Printf("Error getting room: %v", err)
			return
		}

		// Get next word
		nextWord, err := c.hub.wordService.GetRandomWord(ctx, c.roomID, "ru", room.Category)
		if err != nil {
			log.Printf("Error getting next word: %v", err)
			return
		}

		// Set current word in game state
		if err := c.hub.gameService.SetCurrentWord(ctx, c.roomID, nextWord); err != nil {
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
		c.hub.BroadcastToRoom(c.roomID, newWordMsg)
	}
}

func (c *Client) handleVoteStart() {
	log.Printf("Player %d voted to start in room %s", c.user.ID, c.roomID)
}

func (c *Client) handleVotePause() {
	log.Printf("Player %d voted to pause in room %s", c.user.ID, c.roomID)
}

func (c *Client) SendMessage(msg *OutgoingMessage) {
	data, err := json.Marshal(msg)
	if err != nil {
		return
	}

	select {
	case c.send <- data:
	default:
		c.hub.Unregister(c)
	}
}
