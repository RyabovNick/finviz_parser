package telegram

import (
	"log"
	"os"
	"strconv"
)

type Config struct {
	Token string
	Chat  int64
}

func parseTelegramConfig() Config {
	token := os.Getenv("TG_TOKEN")
	if token == "" {
		log.Fatal("ENV: TG_TOKEN not found")
	}

	chat := os.Getenv("CHAT_ID")
	if token == "" {
		log.Fatal("ENV: CHAT_ID not found")
	}

	ch, err := strconv.ParseInt(chat, 10, 64)
	if err != nil {
		log.Fatal("env: cannot convert")
	}

	return Config{
		Token: token,
		Chat:  ch,
	}
}
