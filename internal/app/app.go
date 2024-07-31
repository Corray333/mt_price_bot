package app

import (
	"os"

	"github.com/Corray333/mt_price_bot/internal/storage"
	"github.com/Corray333/mt_price_bot/internal/telegram"
	"github.com/joho/godotenv"
)

type App struct {
	tg *telegram.TelegramClient
}

func New() *App {
	if err := godotenv.Load(os.Args[1]); err != nil {
		panic(err)
	}

	store := storage.New()

	return &App{
		tg: telegram.NewClient(os.Getenv("BOT_TOKEN"), store),
	}
}

func (app *App) Run() {
	app.tg.Run()
}
