package telegram

import (
	"fmt"
	"log/slog"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

const (
	MARK_CHAT_ID = 377742748
)

func (tg *TelegramClient) HandleError(err string, args ...any) {
	if len(args)%2 == 1 {
		args = args[:len(args)-1]
	}
	for i := 0; i < len(args); i += 2 {
		err += fmt.Sprintf("%v=%v", args[i], args[i+1])
	}
	slog.Error(err)

	msg := tgbotapi.NewMessage(MARK_CHAT_ID, err)
	if _, err := tg.bot.Send(msg); err != nil {
		slog.Error("error while handling error: "+err.Error(), "error", err)
	}
}
