package main

import (
	"github.com/FriendlyTornado/DiscordImageBot/pkg/bot"
	"os"

	"github.com/joho/godotenv"
	log "github.com/sirupsen/logrus"
)

func main() {
	// load secrets from .env
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}
	token := os.Getenv("TOKEN")
	bot.Run(token)
	log.Info("DiscordImageBot started")

}
