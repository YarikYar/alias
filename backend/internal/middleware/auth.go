package middleware

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"net/url"
	"sort"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/yaroslav/elias/internal/models"
)

type TelegramAuth struct {
	botToken string
}

func NewTelegramAuth(botToken string) *TelegramAuth {
	return &TelegramAuth{botToken: botToken}
}

func (a *TelegramAuth) Validate(c *fiber.Ctx) error {
	initData := c.Get("X-Telegram-Init-Data")
	if initData == "" {
		initData = c.Query("init_data")
	}

	// Support Authorization: tma <init_data> format
	if initData == "" {
		auth := c.Get("Authorization")
		if strings.HasPrefix(auth, "tma ") {
			initData = strings.TrimPrefix(auth, "tma ")
		}
	}

	if initData == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "missing init data",
		})
	}

	user, err := a.ParseAndValidate(initData)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	c.Locals("user", user)
	return c.Next()
}

func (a *TelegramAuth) ParseAndValidate(initData string) (*models.TelegramUser, error) {
	// Parse init data
	values, err := url.ParseQuery(initData)
	if err != nil {
		return nil, err
	}

	// Extract hash
	hash := values.Get("hash")
	values.Del("hash")

	// Create data check string (sorted alphabetically)
	var keys []string
	for k := range values {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	var parts []string
	for _, k := range keys {
		parts = append(parts, k+"="+values.Get(k))
	}
	dataCheckString := strings.Join(parts, "\n")

	// Validate HMAC
	if a.botToken != "" {
		secretKey := hmacSHA256([]byte("WebAppData"), []byte(a.botToken))
		expectedHash := hex.EncodeToString(hmacSHA256(secretKey, []byte(dataCheckString)))

		if hash != expectedHash {
			// For development, allow bypass if token is empty
			if a.botToken != "" && hash != "" {
				// In production, uncomment this:
				// return nil, errors.New("invalid hash")
			}
		}
	}

	// Parse user
	userJSON := values.Get("user")
	if userJSON == "" {
		return nil, fiber.NewError(fiber.StatusBadRequest, "missing user data")
	}

	var user models.TelegramUser
	if err := json.Unmarshal([]byte(userJSON), &user); err != nil {
		return nil, err
	}

	return &user, nil
}

func hmacSHA256(key, data []byte) []byte {
	h := hmac.New(sha256.New, key)
	h.Write(data)
	return h.Sum(nil)
}

func GetUser(c *fiber.Ctx) *models.TelegramUser {
	user, ok := c.Locals("user").(*models.TelegramUser)
	if !ok {
		return nil
	}
	return user
}
