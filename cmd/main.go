package main

import (
	"github.com/Imm0bilize/gunshot-telegram-notifier/internal/app"
	"github.com/Imm0bilize/gunshot-telegram-notifier/internal/config"
	"log"
)

func main() {
	cfg, err := config.New(".env.public", ".env.private")
	if err != nil {
		log.Fatal(err)
	}

	app.Run(cfg)
}
