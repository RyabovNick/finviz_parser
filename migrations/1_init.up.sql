BEGIN;

CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE last_parse (
  id INT PRIMARY KEY,
  updated_at DATE NOT NULL
);

INSERT INTO last_parse (id, updated_at) VALUES (1, current_date - 100);

CREATE TABLE transactions (
  id UUID DEFAULT uuid_generate_v4() PRIMARY KEY,
  ticker VARCHAR(20) NOT NULL,
  owner VARCHAR(2000) NOT NULL,
  relationship VARCHAR(2000) NOT NULL,
  transaction_date TIMESTAMPTZ NOT NULL,
  transaction_type VARCHAR(2000) NOT NULL,
  cost DOUBLE PRECISION NOT NULL,
  shares BIGINT NOT NULL,
  value BIGINT NOT NULL,
  shares_total BIGINT NOT NULL,
  notification_date TIMESTAMPTZ NOT NULL,
  url VARCHAR(2000) NOT NULL
);

CREATE INDEX ON transactions (notification_date);

COMMIT;