package telegram

import (
	"fmt"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type Connection struct {
	Bot  *tgbotapi.BotAPI
	Chat int64
}

func NewConnection(cfg Config) (*Connection, error) {
	bot, err := tgbotapi.NewBotAPI(cfg.Token)
	if err != nil {
		return nil, fmt.Errorf("error creating bot: %w", err)
	}

	return &Connection{
		Bot:  bot,
		Chat: cfg.Chat,
	}, nil
}

func (c *Connection) SubscriptionNotification() error {
	msg := tgbotapi.NewMessage(c.Chat, fmt.Sprintf(""))

	if _, err := c.Bot.Send(msg); err != nil {
		return fmt.Errorf("error sending message: %w", err)
	}

	return nil
}
