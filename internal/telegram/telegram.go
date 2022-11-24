package telegram

import (
	"context"
	"fmt"
	"strings"

	"github.com/RyabovNick/finviz_parser/internal/insider"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type Storer interface {
	TransactionTypeCount(ctx context.Context) ([]insider.TransactionTypeCount, error)
	TopBuy(ctx context.Context) ([]insider.TotalTransaction, error)
	TopSell(ctx context.Context) ([]insider.TotalTransaction, error)
}

type Connection struct {
	Bot   *tgbotapi.BotAPI
	Chat  int64
	store Storer
}

func New(cfg Config, store Storer) (*Connection, error) {
	bot, err := tgbotapi.NewBotAPI(cfg.Token)
	if err != nil {
		return nil, fmt.Errorf("error creating bot: %w", err)
	}

	return &Connection{
		Bot:   bot,
		Chat:  cfg.Chat,
		store: store,
	}, nil
}

func (c *Connection) Publish(ctx context.Context) error {
	if err := c.transactionTypeCount(ctx); err != nil {
		return fmt.Errorf("error publishing transaction type count: %w", err)
	}

	if err := c.topBuy(ctx); err != nil {
		return fmt.Errorf("error publishing top buy: %w", err)
	}

	if err := c.topSell(ctx); err != nil {
		return fmt.Errorf("error publishing top sell: %w", err)
	}

	return nil
}

func (c *Connection) transactionTypeCount(ctx context.Context) error {
	tt, err := c.store.TransactionTypeCount(ctx)
	if err != nil {
		return fmt.Errorf("error getting transaction type count: %w", err)
	}

	text := make([]string, 0, len(tt)+1)
	text = append(text, "Transaction count and total_value (in $):")

	for _, t := range tt {
		text = append(text, fmt.Sprintf("%s: %d (%.0f)", t.Transaction, t.TransactionCount, t.TotalValue))
	}

	msg := tgbotapi.NewMessage(c.Chat, strings.Join(text, "\n"))

	if _, err := c.Bot.Send(msg); err != nil {
		return fmt.Errorf("error sending message: %w", err)
	}

	return nil
}

func (c *Connection) topBuy(ctx context.Context) error {
	tt, err := c.store.TopBuy(ctx)
	if err != nil {
		return fmt.Errorf("error getting transaction type count: %w", err)
	}

	text := make([]string, 0, len(tt)+1)
	text = append(text, "Top 20 buy:")

	for _, t := range tt {
		text = append(text, fmt.Sprintf("%s: %.0f", t.Ticker, t.TotalValue))
	}

	msg := tgbotapi.NewMessage(c.Chat, strings.Join(text, "\n"))

	if _, err := c.Bot.Send(msg); err != nil {
		return fmt.Errorf("error sending message: %w", err)
	}

	return nil
}

func (c *Connection) topSell(ctx context.Context) error {
	tt, err := c.store.TopSell(ctx)
	if err != nil {
		return fmt.Errorf("error getting transaction type count: %w", err)
	}

	text := make([]string, 0, len(tt)+1)
	text = append(text, "Top 20 sell:")

	for _, t := range tt {
		text = append(text, fmt.Sprintf("%s: %.0f", t.Ticker, t.TotalValue))
	}

	msg := tgbotapi.NewMessage(c.Chat, strings.Join(text, "\n"))

	if _, err := c.Bot.Send(msg); err != nil {
		return fmt.Errorf("error sending message: %w", err)
	}

	return nil
}
