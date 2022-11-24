package main

import (
	"context"

	"github.com/RyabovNick/finviz_parser/internal/insider"
	"github.com/RyabovNick/finviz_parser/internal/store"
	_ "github.com/jackc/pgx/v4/stdlib"
)

func main() {
	ctx := context.Background()

	db, err := store.New(store.Options{
		Host:     "postgres-finviz-parse:5432",
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

	b := insider.New(db)
	txs, err := b.LastDayTransaction()
	if err != nil {
		panic(err)
	}

	if err := b.Save(ctx, txs); err != nil {
		panic(err)
	}
}
