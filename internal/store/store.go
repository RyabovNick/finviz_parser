// Package store содержит взаимодействие с БД
package store

import (
	"context"
	"fmt"
	"strings"
	"time"

	sq "github.com/Masterminds/squirrel"
	"github.com/RyabovNick/finviz_parser/internal/insider"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	pgsq                  = sq.StatementBuilder.PlaceholderFormat(sq.Dollar)
	ErrAlreadyParsedToday = fmt.Errorf("already parsed today")
)

type Options struct {
	Host     string
	Database string
	Username string
	Password string
	MaxPool  int
	MinPool  int
}

func (o Options) String() string {
	port := "5432"

	hp := strings.Split(o.Host, ":")
	if len(hp) == 2 {
		o.Host = hp[0]
		port = hp[1]
	}

	return fmt.Sprintf("host=%s port=%s dbname=%s user=%s password=%s sslmode=disable pool_min_conns=%d pool_max_conns=%d", o.Host, port, o.Database, o.Username, o.Password, o.MinPool, o.MaxPool)
}

type Store struct {
	pool *pgxpool.Pool
}

// New creates connection.
func New(ctx context.Context, o Options) (*Store, error) {
	pool, err := pgxpool.New(ctx, o.String())
	if err != nil {
		return nil, fmt.Errorf("failed connection: %w", err)
	}

	// ping
	ctx, cancel := context.WithTimeout(ctx, time.Second*5)
	defer cancel()
	if err := pool.Ping(ctx); err != nil {
		return nil, fmt.Errorf("failed ping: %w", err)
	}
	return &Store{pool: pool}, nil
}

func (s *Store) Close() {
	s.pool.Close()
}

func (s *Store) lastParse(ctx context.Context) (time.Time, error) {
	var t time.Time
	if err := s.pool.QueryRow(ctx, `
		SELECT updated_at
		FROM last_parse
		WHERE id = 1;
	`).Scan(&t); err != nil {
		return time.Time{}, fmt.Errorf("failed select last parse: %w", err)
	}

	return t, nil
}

// updateLastParse sets last parse as yesterday.
func (s *Store) updateLastParse(ctx context.Context) error {
	if _, err := s.pool.Exec(ctx, `
		UPDATE last_parse
		SET updated_at = current_date - 1
		WHERE id = 1;
	`); err != nil {
		return fmt.Errorf("failed update last parse: %w", err)
	}

	return nil
}

func (s *Store) InsertTransactions(ctx context.Context, tr insider.Transactions) error {
	lp, err := s.lastParse(ctx)
	if err != nil {
		return fmt.Errorf("failed get last parse: %w", err)
	}

	query := pgsq.Insert("transactions").Columns("ticker", "owner", "relationship",
		"transaction_date", "transaction_type", "cost", "shares", "value",
		"shares_total", "notification_date", "url")

	for _, t := range tr {
		// skip if already parsed today
		if t.NotificationDate.Format("2006-01-02") == lp.Format("2006-01-02") {
			return ErrAlreadyParsedToday
		}

		query = query.Values(t.Ticker, t.Owner, t.Relationship, t.TransactionDate,
			t.Transaction, t.Cost, t.Shares, t.Value, t.SharesTotal, t.SEC.NotificationDate, t.SEC.URL)
	}

	sql, args, err := query.ToSql()
	if err != nil {
		return fmt.Errorf("candles insert to sql: %w", err)
	}

	if _, err := s.pool.Exec(ctx, sql, args...); err != nil {
		return fmt.Errorf("candles insert exec: %w", err)
	}

	if err := s.updateLastParse(ctx); err != nil {
		return fmt.Errorf("failed update last parse: %w", err)
	}

	return nil
}

func (s *Store) TransactionTypeCount(ctx context.Context) ([]insider.TransactionTypeCount, error) {
	rows, _ := s.pool.Query(ctx, `
		SELECT transaction_type, count(*) as transaction_count, sum(value) as total_value
		FROM transactions
		WHERE notification_date::date = current_date - 1
		GROUP BY transaction_type
		ORDER BY transaction_type;
	`)
	tc, err := pgx.CollectRows(rows, pgx.RowToStructByName[insider.TransactionTypeCount])
	if err != nil {
		return nil, fmt.Errorf("failed select transaction type count: %w", err)
	}

	return tc, nil
}

func (s *Store) RelationshipCount(ctx context.Context) ([]insider.RelationshipCount, error) {
	rows, _ := s.pool.Query(ctx, `
		SELECT relationship, transaction_type, count(*) as transaction_count, sum(value) as total_value
		FROM transactions
		WHERE notification_date::date = current_date - 1
		GROUP BY relationship, transaction_type
		ORDER BY total_value DESC;
	`)
	rc, err := pgx.CollectRows(rows, pgx.RowToStructByName[insider.RelationshipCount])
	if err != nil {
		return nil, fmt.Errorf("failed select relationship count: %w", err)
	}

	return rc, nil
}

func (s *Store) TopBuy(ctx context.Context) ([]insider.TotalTransaction, error) {
	rows, _ := s.pool.Query(ctx, `
		WITH sale AS (
			SELECT ticker, sum(value) as total_value
			FROM transactions
			WHERE notification_date::date = current_date - 1
				AND transaction_type = 'Sale'
			GROUP BY ticker
		), buy AS (
				SELECT ticker, sum(value) as total_value
				FROM transactions
				WHERE notification_date::date = current_date - 1
					AND transaction_type = 'Buy'
				GROUP BY ticker
		)
		SELECT
					CASE WHEN sale.ticker IS NULL THEN buy.ticker ELSE sale.ticker END as ticker,
					CASE WHEN sale.total_value IS NULL THEN buy.total_value ELSE
								CASE WHEN buy.total_value IS NULL THEN -sale.total_value ELSE buy.total_value - sale.total_value END END AS total_value
		FROM sale
		FULL OUTER JOIN buy ON sale.ticker = buy.ticker
		ORDER BY total_value DESC
		LIMIT 20;
	`)
	tt, err := pgx.CollectRows(rows, pgx.RowToStructByName[insider.TotalTransaction])
	if err != nil {
		return nil, fmt.Errorf("failed select top sell: %w", err)
	}

	return tt, nil
}

func (s *Store) TopSell(ctx context.Context) ([]insider.TotalTransaction, error) {
	rows, _ := s.pool.Query(ctx, `
		WITH sale AS (
			SELECT ticker, sum(value) as total_value
			FROM transactions
			WHERE notification_date::date = current_date - 1
				AND transaction_type = 'Sale'
			GROUP BY ticker
		), buy AS (
				SELECT ticker, sum(value) as total_value
				FROM transactions
				WHERE notification_date::date = current_date - 1
					AND transaction_type = 'Buy'
				GROUP BY ticker
		)
		SELECT
					CASE WHEN sale.ticker IS NULL THEN buy.ticker ELSE sale.ticker END as ticker,
					CASE WHEN sale.total_value IS NULL THEN buy.total_value ELSE
								CASE WHEN buy.total_value IS NULL THEN -sale.total_value ELSE buy.total_value - sale.total_value END END AS total_value
		FROM sale
		FULL OUTER JOIN buy ON sale.ticker = buy.ticker
		ORDER BY total_value ASC
		LIMIT 20;
	`)
	tc, err := pgx.CollectRows(rows, pgx.RowToStructByName[insider.TotalTransaction])
	if err != nil {
		return nil, fmt.Errorf("failed select top sell: %w", err)
	}

	return tc, nil
}

func (s *Store) SaleTicker(ctx context.Context) (insider.Tickers, error) {
	rows, _ := s.pool.Query(ctx, `
		SELECT DISTINCT ticker
		FROM transactions
		WHERE notification_date::date = current_date - 1
			AND transaction_type = 'Sale';
	`)
	t, err := pgx.CollectRows(rows, pgx.RowTo[string])
	if err != nil {
		return nil, fmt.Errorf("failed select sale ticker: %w", err)
	}

	return t, nil
}

func (s *Store) BuyTicker(ctx context.Context) (insider.Tickers, error) {
	rows, _ := s.pool.Query(ctx, `
		SELECT DISTINCT ticker
		FROM transactions
		WHERE notification_date::date = current_date - 1
			AND transaction_type = 'Buy';
	`)
	t, err := pgx.CollectRows(rows, pgx.RowTo[string])
	if err != nil {
		return nil, fmt.Errorf("failed select buy ticker: %w", err)
	}

	return t, nil
}
