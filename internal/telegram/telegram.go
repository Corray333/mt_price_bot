package telegram

import (
	"database/sql"
	"log"

	"github.com/Corray333/mt_price_bot/internal/types"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type Storage interface {
	UpdateUser(user *types.User) error
	CreateUser(user *types.User) error
	GetAllUsers() ([]types.User, error)
	GetUserByID(user_id int64) (*types.User, error)
}

const (
	StateWaitingFIO = iota + 1
	StateWaitingEmail
	StateWaitingDirection
	StateWaitingGroup
)

const (
	StateNothing = iota + 1
	StateWaitingUserTypePick
	StateWaitingMessageText
	StateWaitingMessageAttachment
	StateWaitingSending
)

type TelegramClient struct {
	bot   *tgbotapi.BotAPI
	store Storage
}

var messages = []string{
	"Не понимаю, чего ты хочешь😅",
	"Прости, я тебя не понял🤔",
	"Что-то пошло не так, попробуй снова🙏",
	"Не могу разобраться, попробуй иначе😉",
	"Похоже, я тебя не понимаю😕",
	"Давай попробуем еще раз, я тебя не понял😊",
	"Может быть, я что-то упустил. Попробуй ещё раз😌",
	"Извини, я тебя не понял. Попробуй сформулировать иначе🤷‍♂️",
	"Я не совсем понял твоё действие. Попробуй что-то другое🙃",
	"Не могу распознать твой запрос. Попробуй снова🧐",
}

func NewClient(token string, store Storage) *TelegramClient {
	bot, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		log.Fatal("Failed to create bot: ", err)
	}

	bot.Debug = true

	return &TelegramClient{
		bot:   bot,
		store: store,
	}
}

func (tg *TelegramClient) Run() {
	defer func() {
		if r := recover(); r != nil {
			tg.HandleError("panic: " + r.(string))
		}
	}()
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := tg.bot.GetUpdatesChan(u)

	for update := range updates {
		if update.Message == nil && update.CallbackQuery == nil {
			continue
		}

		user, err := tg.store.GetUserByID(update.Message.Chat.ID)
		if err != nil {
			tg.HandleError("error while getting user from db: "+err.Error(), "update_id", update.UpdateID)
			return
		}

		switch {
		case user.IsAdmin:
			tg.handleAdminUpdate(user, update)
			continue
		default:
			tg.handleUserUpdate(user, update)
			continue
		}

	}
}

func (tg *TelegramClient) handleUserUpdate(user *types.User, update tgbotapi.Update) {

	switch {
	}

	if err := tg.store.UpdateUser(user); err != nil {
		tg.HandleError("error while updating user: "+err.Error(), "update_id", update.UpdateID)
	}
}

func (tg *TelegramClient) handleAdminUpdate(user *types.User, update tgbotapi.Update) {

}

func (tg *TelegramClient) sendWelcomeMessage(chatID int64) {
	_, err := tg.store.GetUserByID(chatID)
	if err != nil {
		if err == sql.ErrNoRows {
			button := tgbotapi.NewKeyboardButtonContact("Отправить контакт")
			keyboard := tgbotapi.NewReplyKeyboard([]tgbotapi.KeyboardButton{button})
			msg := tgbotapi.NewMessage(chatID, "Привет! Это команда Incetro.\nЧтобы зарегистрироваться на стажировку, поделись своим контактом🤙")
			msg.ReplyMarkup = keyboard
			if _, err := tg.bot.Send(msg); err != nil {
				tg.HandleError("error while sending message: "+err.Error(), "chat_id", chatID)
				return
			}
			return
		}
		tg.HandleError("error while getting user from db: "+err.Error(), "chat_id", chatID)
		return
	}
	msg := tgbotapi.NewMessage(chatID, "Прости, но я не понимаю, что ты от меня хочешь😥")
	if _, err := tg.bot.Send(msg); err != nil {
		tg.HandleError("error while sending message: "+err.Error(), "chat_id", chatID)
		return
	}

}
