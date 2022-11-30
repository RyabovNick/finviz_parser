package telegram

import (
	"context"
	"fmt"
	"strings"

	"github.com/RyabovNick/finviz_parser/internal/insider"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

const (
	ParseModeHTML = "HTML"
)

type Storer interface {
	TransactionTypeCount(ctx context.Context) ([]insider.TransactionTypeCount, error)

	TopBuy(ctx context.Context) ([]insider.TotalTransaction, error)
	TopSell(ctx context.Context) ([]insider.TotalTransaction, error)

	BuyTicker(ctx context.Context) (insider.Tickers, error)
	SaleTicker(ctx context.Context) (insider.Tickers, error)
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

	if len(tt) == 0 {
		return fmt.Errorf("transaction type count is empty")
	}

	text := make([]string, 0, len(tt)+1)
	text = append(text, "<b>Transaction count and total_value (in $):</b>")

	for _, t := range tt {
		text = append(text, fmt.Sprintf("%s: %d (%.0f)", t.Transaction, t.TransactionCount, t.TotalValue))
	}

	msg := tgbotapi.NewMessage(c.Chat, strings.Join(text, "\n"))
	msg.ParseMode = ParseModeHTML

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

	if len(tt) == 0 {
		return fmt.Errorf("top buy is empty")
	}

	text := make([]string, 0, len(tt)+2)
	text = append(text, "<b>Top 20 buy:</b>")

	for _, t := range tt {
		text = append(text, fmt.Sprintf("%s: %.0f", t.FinvizTicker(), t.TotalValue))
	}

	tick, err := c.store.BuyTicker(ctx)
	if err != nil {
		return fmt.Errorf("error getting tickers: %w", err)
	}

	text = append(text, tick.Finviz())

	msg := tgbotapi.NewMessage(c.Chat, strings.Join(text, "\n"))
	msg.ParseMode = ParseModeHTML

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

	if len(tt) == 0 {
		return fmt.Errorf("top sell is empty")
	}

	text := make([]string, 0, len(tt)+1)
	text = append(text, "<b>Top 20 sell:</b>")

	for _, t := range tt {
		text = append(text, fmt.Sprintf("%s: %.0f", t.FinvizTicker(), t.TotalValue))
	}

	tick, err := c.store.SaleTicker(ctx)
	if err != nil {
		return fmt.Errorf("error getting tickers: %w", err)
	}

	text = append(text, tick.Finviz())

	msg := tgbotapi.NewMessage(c.Chat, strings.Join(text, "\n"))
	msg.ParseMode = ParseModeHTML

	if _, err := c.Bot.Send(msg); err != nil {
		return fmt.Errorf("error sending message: %w", err)
	}

	return nil
}
