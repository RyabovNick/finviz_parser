package main

import (
	"context"

	"github.com/RyabovNick/finviz_parser/internal/insider"
	"github.com/RyabovNick/finviz_parser/internal/store"
	"github.com/RyabovNick/finviz_parser/internal/telegram"
	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load(".env")
	if err != nil {
		panic(err)
	}

	ctx := context.Background()

	db, err := store.New(ctx, store.Options{
		Host:     "localhost:5432",
		Database: "test",
		Username: "postgres",
		Password: "pass",
		MaxPool:  10,
		MinPool:  2,
	})
	if err != nil {
		panic(err)
	}
	defer db.Close()

	telegram, err := telegram.New(telegram.ParseTelegramConfig(), db)
	if err != nil {
		panic(err)
	}

	b := insider.New(db)
	txs, err := b.LastDayTransaction()
	if err != nil {
		panic(err)
	}

	if err := b.Save(ctx, txs); err != nil {
		panic(err)
	}

	if err := telegram.Publish(ctx); err != nil {
		panic(err)
	}
}
