package telegram

import (
	"database/sql"
	"fmt"
	"net/mail"
	"regexp"
	"slices"
	"strconv"

	"github.com/Corray333/mt_price_bot/internal/storage"
	"github.com/Corray333/mt_price_bot/internal/types"
	"github.com/Corray333/mt_price_bot/internal/utils"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

const (
	StateWaitingFIO = iota + 1
	StateWaitingEmail
	StateWaitingPhone
	StateWaitingOrgName
	StateWaitingOrgsNumber
	StateDone
)

func (tg *TelegramClient) sendWelcomeMessage(update tgbotapi.Update) {
	_, err := tg.store.GetUserByID(update.FromChat().ID)
	if err != nil && err != sql.ErrNoRows {
		tg.HandleError("error while getting user: "+err.Error(), "update", update.UpdateID)
		return
	}
	if err == sql.ErrNoRows {
		if err := tg.store.CreateUser(&types.User{
			ID:       update.FromChat().ID,
			Username: update.Message.From.UserName,
			IsAdmin:  slices.Contains(storage.Admins, update.Message.From.UserName),
		}); err != nil {
			tg.HandleError("error while creating user: "+err.Error(), "update", update.UpdateID)
			msg := tgbotapi.NewMessage(update.FromChat().ID, storage.Messages[storage.MsgError])
			if _, err := tg.bot.Send(msg); err != nil {
				tg.HandleError("error while sending message: "+err.Error(), "update", update.UpdateID)
				return
			}
			return
		}
	}

	if slices.Contains(storage.Admins, update.Message.From.UserName) {
		msg := tgbotapi.NewMessage(update.FromChat().ID, "Добро пожаловать в админку. Все заявки будут приходить в чат. Чтобы обновить прайс, отправьте файл в чате.")
		msg.ReplyMarkup = tgbotapi.NewReplyKeyboard(
			tgbotapi.NewKeyboardButtonRow(
				tgbotapi.NewKeyboardButton("Обновить тексты сообщений"),
			),
		)
		if _, err := tg.bot.Send(msg); err != nil {
			tg.HandleError("error while sending message: "+err.Error(), "update", update.UpdateID)
			return
		}
		return
	}

	msg := tgbotapi.NewMessage(update.FromChat().ID, storage.Messages[storage.MsgWelcome])
	msg.ReplyMarkup = tgbotapi.NewReplyKeyboard(
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton(storage.Messages[storage.ButtonForm]),
			tgbotapi.NewKeyboardButton(storage.Messages[storage.ButtonPrice]),
		),
	)
	if _, err := tg.bot.Send(msg); err != nil {
		tg.HandleError("error while sending message: "+err.Error(), "update", update.UpdateID)
		return
	}
}

func (tg *TelegramClient) sendPrice(update tgbotapi.Update) {
	fileName, err := utils.FindFileWithKeyword("price")
	if err != nil {
		tg.HandleError("error while finding file: "+err.Error(), "update", update.UpdateID)
		msg := tgbotapi.NewMessage(update.FromChat().ID, storage.Messages[storage.MsgError])
		if _, err := tg.bot.Send(msg); err != nil {
			tg.HandleError("error while sending message: "+err.Error(), "update", update.UpdateID)
			return
		}
		return
	}
	msg := tgbotapi.NewDocument(update.FromChat().ID, tgbotapi.FilePath("../files/"+fileName))
	msg.Caption = storage.Messages[storage.MsgPrice]
	if _, err := tg.bot.Send(msg); err != nil {
		tg.HandleError("error while sending message: "+err.Error(), "update", update.UpdateID)
		return
	}
}

func (tg *TelegramClient) sendForm(user *types.User, update tgbotapi.Update) {
	msg := tgbotapi.NewMessage(update.FromChat().ID, storage.Messages[storage.MsgAskFIO])
	msg.ReplyMarkup = tgbotapi.NewRemoveKeyboard(true)
	if _, err := tg.bot.Send(msg); err != nil {
		tg.HandleError("error while sending message: "+err.Error(), "update", update.UpdateID)
		return
	}
	user.State = StateWaitingFIO
	if err := tg.store.UpdateUser(user); err != nil {
		tg.HandleError("error while updating user: "+err.Error(), "update", update.UpdateID)
		return
	}
}

func (tg *TelegramClient) handleInputFIO(user *types.User, update tgbotapi.Update) {
	re := regexp.MustCompile(`^([A-ZА-Я][a-zа-яA-ZА-Я]*\s){1,2}[A-ZА-Я][a-zа-яA-ZА-Я]*$`)
	if !re.MatchString(update.Message.Text) {
		msg := tgbotapi.NewMessage(update.FromChat().ID, "Пожалуйста, введи только имя и фамилию)")
		if _, err := tg.bot.Send(msg); err != nil {
			tg.HandleError("error while sending message: "+err.Error(), "update", update.UpdateID)
			return
		}
		return
	}
	user.FIO = update.Message.Text
	user.State = StateWaitingEmail

	msg := tgbotapi.NewMessage(update.FromChat().ID, storage.Messages[storage.MsgAskEmail])
	if _, err := tg.bot.Send(msg); err != nil {
		tg.HandleError("error while sending message: "+err.Error(), "update", update.UpdateID)
		return
	}

}

func (tg *TelegramClient) handleInputEmail(user *types.User, update tgbotapi.Update) {
	if _, err := mail.ParseAddress(update.Message.Text); err != nil {
		msg := tgbotapi.NewMessage(update.FromChat().ID, "Введите корректную почту")
		if _, err := tg.bot.Send(msg); err != nil {
			tg.HandleError("error while sending message: "+err.Error(), "update", update.UpdateID)
			return
		}
		return
	}

	msg := tgbotapi.NewMessage(update.FromChat().ID, storage.Messages[storage.MsgAskPhone])
	if _, err := tg.bot.Send(msg); err != nil {
		tg.HandleError("error while sending message: "+err.Error(), "update", update.UpdateID)
		return
	}

	user.Email = update.Message.Text
	user.State = StateWaitingPhone

}

func (tg *TelegramClient) handleInputPhone(user *types.User, update tgbotapi.Update) {
	phoneRegex := `^(\+?\d{1,3})? ?(\(?\d{1,4}\)?)? ?[\d\s-]{3,15}$`
	re := regexp.MustCompile(phoneRegex)
	fmt.Println(update.Message.Text)
	fmt.Println(re.MatchString(update.Message.Text))
	if ok := re.MatchString(update.Message.Text); !ok {
		msg := tgbotapi.NewMessage(update.FromChat().ID, "Введите корректный номер телефона")
		if _, err := tg.bot.Send(msg); err != nil {
			tg.HandleError("error while sending message: "+err.Error(), "update", update.UpdateID)
			return
		}
		return
	}

	msg := tgbotapi.NewMessage(update.FromChat().ID, storage.Messages[storage.MsgAskOrgName])
	msg.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(storage.Messages[storage.ButtonNoOrg], "no_org"),
		),
	)
	if _, err := tg.bot.Send(msg); err != nil {
		tg.HandleError("error while sending message: "+err.Error(), "update", update.UpdateID)
		return
	}

	user.Phone = update.Message.Text
	user.State = StateWaitingOrgName
}

func (tg *TelegramClient) handleInputOrgName(user *types.User, update tgbotapi.Update) {

	if update.CallbackQuery != nil {
		cb := tgbotapi.NewCallback(update.CallbackQuery.ID, "")
		if _, err := tg.bot.Request(cb); err != nil {
			tg.HandleError("error while sending message: "+err.Error(), "update", update.UpdateID)
			return
		}
		del := tgbotapi.NewDeleteMessage(update.FromChat().ID, update.CallbackQuery.Message.MessageID)
		if _, err := tg.bot.Request(del); err != nil {
			tg.HandleError("error while deleting message: "+err.Error(), "update", update.UpdateID)
			return
		}

		user.Org = "-"
		user.OrgNumber = 0
		user.State = StateDone

		if err := tg.store.UpdateUser(user); err != nil {
			tg.HandleError("error while updating user: "+err.Error(), "update", update.UpdateID)
			return
		}

		msg := tgbotapi.NewMessage(update.FromChat().ID, storage.Messages[storage.MsgAccepted])
		msg.ReplyMarkup = tgbotapi.NewReplyKeyboard(
			tgbotapi.NewKeyboardButtonRow(
				// tgbotapi.NewKeyboardButton(storage.Messages[storage.ButtonForm]),
				tgbotapi.NewKeyboardButton(storage.Messages[storage.ButtonPrice]),
			),
		)
		if _, err := tg.bot.Send(msg); err != nil {
			tg.HandleError("error while sending message: "+err.Error(), "update", update.UpdateID)
			return
		}

		admins, err := tg.store.GetAllAdmins()
		if err != nil {
			tg.HandleError("error while getting admins: "+err.Error(), "update", update.UpdateID)
			return
		}

		text := fmt.Sprintf("**Новая заявка**\nФИО:  [%s](%s)\nТелефон: %s\nКонтактная почта: %s\nКоличество точек: %d\nНазвание точки: %s\n", user.FIO, "t.me/"+user.Username, user.Phone, user.Email, user.OrgNumber, user.Org)

		for _, admin := range admins {
			msg := tgbotapi.NewMessage(admin.ID, text)
			msg.ParseMode = tgbotapi.ModeMarkdown
			if _, err := tg.bot.Send(msg); err != nil {
				tg.HandleError("error while sending message: "+err.Error(), "update", update.UpdateID)
				return
			}
		}
		return
	}

	msg := tgbotapi.NewMessage(update.FromChat().ID, storage.Messages[storage.MsgAskOrgsNumber])
	if _, err := tg.bot.Send(msg); err != nil {
		tg.HandleError("error while sending message: "+err.Error(), "chat_id", "update", update.UpdateID)
		return
	}

	user.Org = update.Message.Text
	user.State = StateWaitingOrgsNumber
}

func (tg *TelegramClient) handleInputOrgNumber(user *types.User, update tgbotapi.Update) {

	num, err := strconv.Atoi(update.Message.Text)
	if err != nil {
		msg := tgbotapi.NewMessage(update.FromChat().ID, "Введите корректное количество точек")
		if _, err := tg.bot.Send(msg); err != nil {
			tg.HandleError("error while sending message: "+err.Error(), "update", update.UpdateID)
			return
		}
		return
	}

	user.OrgNumber = num
	user.State = StateDone

	if err := tg.store.UpdateUser(user); err != nil {
		tg.HandleError("error while updating user: "+err.Error(), "update", update.UpdateID)
		return
	}

	msg := tgbotapi.NewMessage(update.FromChat().ID, storage.Messages[storage.MsgAccepted])
	msg.ReplyMarkup = tgbotapi.NewReplyKeyboard(
		tgbotapi.NewKeyboardButtonRow(
			// tgbotapi.NewKeyboardButton(storage.Messages[storage.ButtonForm]),
			tgbotapi.NewKeyboardButton(storage.Messages[storage.ButtonPrice]),
		),
	)
	if _, err := tg.bot.Send(msg); err != nil {
		tg.HandleError("error while sending message: "+err.Error(), "update", update.UpdateID)
		return
	}

	admins, err := tg.store.GetAllAdmins()
	if err != nil {
		tg.HandleError("error while getting admins: "+err.Error(), "update", update.UpdateID)
		return
	}

	text := fmt.Sprintf("**Новая заявка**\nФИО:  [%s](%s)\nТелефон: %s\nКонтактная почта: %s\nКоличество точек: %d\nНазвание точки: %s\n", user.FIO, "t.me/"+user.Username, user.Phone, user.Email, user.OrgNumber, user.Org)

	for _, admin := range admins {
		msg := tgbotapi.NewMessage(admin.ID, text)
		msg.ParseMode = tgbotapi.ModeMarkdown
		if _, err := tg.bot.Send(msg); err != nil {
			tg.HandleError("error while sending message: "+err.Error(), "update", update.UpdateID)
			return
		}
	}
}
