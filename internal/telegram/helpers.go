package telegram

import (
	"encoding/json"
	"fmt"
	"log/slog"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

const (
	MARK_CHAT_ID = 377742748
)

func (tg *TelegramClient) HandleError(errMsg string, args ...any) {
	errObj := map[string]interface{}{
		"error": errMsg,
	}
	if len(args)%2 == 1 {
		args = args[:len(args)-1]
	}
	for i := 0; i < len(args); i += 2 {
		errObj[fmt.Sprintf("%v", args[i])] = args[i+1]
		errMsg += fmt.Sprintf("%v=%v", args[i], args[i+1])
	}
	slog.Error(errMsg)

	marshalled, err := json.MarshalIndent(errObj, "", "\t")
	if err != nil {
		slog.Error("error while marshalling error object: " + err.Error())
		return
	}

	msg := tgbotapi.NewMessage(MARK_CHAT_ID, fmt.Sprintf("```%s```", string(marshalled)))
	msg.ParseMode = tgbotapi.ModeMarkdown
	if _, err := tg.bot.Send(msg); err != nil {
		slog.Error("error while handling error: "+err.Error(), "error", err)
	}
}
