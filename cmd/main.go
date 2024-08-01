package main

import (
	"log"
	"os"

	"github.com/Corray333/mt_price_bot/internal/app"
	"github.com/Corray333/mt_price_bot/internal/gsheets"
	"github.com/joho/godotenv"
)

func main() {
	godotenv.Load(os.Args[1])
	if err := gsheets.UpdateMessages(); err != nil {
		log.Fatal(err)
	}

	app.New().Run()
}
