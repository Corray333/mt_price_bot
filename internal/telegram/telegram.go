package telegram

import (
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"github.com/Corray333/mt_price_bot/internal/gsheets"
	"github.com/Corray333/mt_price_bot/internal/storage"
	"github.com/Corray333/mt_price_bot/internal/types"
	"github.com/Corray333/mt_price_bot/internal/utils"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type Storage interface {
	UpdateUser(user *types.User) error
	CreateUser(user *types.User) error
	GetUserByID(user_id int64) (*types.User, error)
	GetAllAdmins() ([]*types.User, error)
}

type TelegramClient struct {
	bot   *tgbotapi.BotAPI
	store Storage
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

		if update.Message != nil {
			if update.Message.IsCommand() {
				if update.Message.Command() == "start" {
					tg.sendWelcomeMessage(update)
					continue
				}

			}
		}
		user, err := tg.store.GetUserByID(update.FromChat().ID)
		if err != nil {
			tg.HandleError("error while getting user from db: "+err.Error(), "update_id", update.UpdateID)
			continue
		}

		switch {
		case user.IsAdmin:
			tg.handleAdminUpdate(update)
			continue
		default:
			tg.handleUserUpdate(user, update)
			continue
		}

	}
}

func (tg *TelegramClient) handleUserUpdate(user *types.User, update tgbotapi.Update) {

	if update.Message != nil {
		if update.Message.Text == storage.Messages[storage.ButtonPrice] {
			tg.sendPrice(user, update)
			if err := tg.store.UpdateUser(user); err != nil {
				tg.HandleError("error while updating user: "+err.Error(), "update_id", update.UpdateID)
			}
			return
		} else if update.Message.Text == storage.Messages[storage.ButtonForm] {
			tg.sendForm(user, update)
			return
		}
	}

	switch user.State {
	case StateWaitingFIO:
		tg.handleInputFIO(user, update)
	case StateWaitingPhoneForOrder:
		tg.handleInputPhoneForOrder(user, update)
	case StateWaitingEmail:
		tg.handleInputEmail(user, update)
	case StateWaitingPhone:
		tg.handleInputPhone(user, update)
	case StateWaitingOrgName:
		tg.handleInputOrgName(user, update)
	case StateWaitingOrgsNumber:
		tg.handleInputOrgNumber(user, update)
	}

	if err := tg.store.UpdateUser(user); err != nil {
		tg.HandleError("error while updating user: "+err.Error(), "update_id", update.UpdateID)
	}
}

func (tg *TelegramClient) handleAdminUpdate(update tgbotapi.Update) {
	if update.Message != nil {
		if update.Message.IsCommand() {
			if update.Message.Command() == "start" {
				tg.sendWelcomeMessage(update)
				return
			}
		}
		if update.Message.Text == "Обновить тексты сообщений" {
			gsheets.UpdateMessages()
			msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Данные обновлены")
			if _, err := tg.bot.Send(msg); err != nil {
				tg.HandleError("error while sending message: "+err.Error(), "update_id", update.UpdateID)
				return
			}
			return
		}
		if update.Message.Document != nil {
			tg.handleNewPrice(update)
			return
		}
	}
}

func (tg *TelegramClient) handleNewPrice(update tgbotapi.Update) {
	if err := utils.RemoveFilesWithKeyword("price"); err != nil {
		tg.HandleError("error while removing files: "+err.Error(), "update_id", update.UpdateID)
		return
	}
	doc := update.Message.Document
	fileID := doc.FileID

	file, err := tg.bot.GetFile(tgbotapi.FileConfig{FileID: fileID})
	if err != nil {
		tg.HandleError("error while getting file: "+err.Error(), "update_id", update.UpdateID)
		return
	}

	fileURL := file.Link(tg.bot.Token)

	response, err := http.Get(fileURL)
	if err != nil {
		tg.HandleError("error while getting file: "+err.Error(), "update_id", update.UpdateID)
		return
	}
	defer response.Body.Close()

	extension := filepath.Ext(doc.FileName)
	newFileName := "../files/price" + extension

	out, err := os.Create(newFileName)
	if err != nil {
		tg.HandleError("error while creating file: "+err.Error(), "update_id", update.UpdateID)
		return
	}
	defer out.Close()

	_, err = io.Copy(out, response.Body)
	if err != nil {
		tg.HandleError("error while copying file: "+err.Error(), "update_id", update.UpdateID)
		return
	}

	msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Прайс обновлен")
	if _, err := tg.bot.Send(msg); err != nil {
		tg.HandleError("error while sending message: "+err.Error(), "update_id", update.UpdateID)
		return
	}
}
