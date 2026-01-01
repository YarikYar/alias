package bot

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"
)

type Bot struct {
	token  string
	appURL string
}

func New(token, appURL string) *Bot {
	return &Bot{
		token:  token,
		appURL: appURL,
	}
}

type Update struct {
	UpdateID int      `json:"update_id"`
	Message  *Message `json:"message,omitempty"`
}

type Message struct {
	Chat      Chat   `json:"chat"`
	Text      string `json:"text"`
	MessageID int    `json:"message_id"`
}

type Chat struct {
	ID int64 `json:"id"`
}

func (b *Bot) Start() {
	log.Println("Bot polling started...")
	offset := 0

	for {
		updates, err := b.getUpdates(offset)
		if err != nil {
			log.Printf("Error getting updates: %v", err)
			time.Sleep(3 * time.Second)
			continue
		}

		for _, update := range updates {
			offset = update.UpdateID + 1
			b.handleUpdate(update)
		}

		time.Sleep(1 * time.Second)
	}
}

func (b *Bot) getUpdates(offset int) ([]Update, error) {
	url := fmt.Sprintf("https://api.telegram.org/bot%s/getUpdates?offset=%d&timeout=30", b.token, offset)
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result struct {
		Ok     bool     `json:"ok"`
		Result []Update `json:"result"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return result.Result, nil
}

func (b *Bot) handleUpdate(update Update) {
	if update.Message == nil {
		return
	}

	if update.Message.Text == "/start" {
		b.sendWebAppButton(update.Message.Chat.ID)
	}
}

func (b *Bot) sendWebAppButton(chatID int64) {
	url := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", b.token)

	payload := map[string]interface{}{
		"chat_id": chatID,
		"text":    "🎮 Добро пожаловать в Alias!\n\nНажмите кнопку ниже, чтобы начать игру:",
		"reply_markup": map[string]interface{}{
			"inline_keyboard": [][]map[string]interface{}{
				{
					{
						"text": "🎯 Играть",
						"web_app": map[string]string{
							"url": b.appURL,
						},
					},
				},
			},
		},
	}

	data, _ := json.Marshal(payload)
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(data))
	if err != nil {
		log.Printf("Error sending message: %v", err)
		return
	}
	defer resp.Body.Close()
}
