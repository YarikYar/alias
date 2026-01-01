package handlers

import (
	"github.com/gofiber/contrib/websocket"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/yaroslav/elias/internal/middleware"
	"github.com/yaroslav/elias/internal/ws"
)

type WSHandler struct {
	hub  *ws.Hub
	auth *middleware.TelegramAuth
}

func NewWSHandler(hub *ws.Hub, auth *middleware.TelegramAuth) *WSHandler {
	return &WSHandler{hub: hub, auth: auth}
}

func (h *WSHandler) HandleWebSocket(c *fiber.Ctx) error {
	// Upgrade check
	if !websocket.IsWebSocketUpgrade(c) {
		return fiber.ErrUpgradeRequired
	}

	roomID, err := uuid.Parse(c.Params("room"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid room id"})
	}

	// Get init data from query
	initData := c.Query("init_data")
	if initData == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "missing init data"})
	}

	user, err := h.auth.ParseAndValidate(initData)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": err.Error()})
	}

	return websocket.New(func(conn *websocket.Conn) {
		client := ws.NewClient(h.hub, conn, roomID, user)
		h.hub.Register(client)

		go client.WritePump()
		client.ReadPump()
	})(c)
}
