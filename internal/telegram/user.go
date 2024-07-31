package telegram

import (
	"fmt"
	"net/mail"
	"regexp"
	"slices"
	"strconv"

	"github.com/Corray333/mt_price_bot/internal/types"
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
	if err := tg.store.CreateUser(&types.User{
		ID:       update.FromChat().ID,
		Username: update.Message.From.UserName,
		IsAdmin:  slices.Contains(admins, update.Message.From.UserName),
	}); err != nil {
		tg.HandleError("error while creating user: "+err.Error(), "update", update.UpdateID)
		msg := tgbotapi.NewMessage(update.FromChat().ID, messages[MsgError])
		if _, err := tg.bot.Send(msg); err != nil {
			tg.HandleError("error while sending message: "+err.Error(), "update", update.UpdateID)
			return
		}
		return
	}

	msg := tgbotapi.NewMessage(update.FromChat().ID, messages[MsgWelcome])
	msg.ReplyMarkup = tgbotapi.NewReplyKeyboard(
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton(messages[ButtonForm]),
			tgbotapi.NewKeyboardButton(messages[ButtonPrice]),
		),
	)
	if _, err := tg.bot.Send(msg); err != nil {
		tg.HandleError("error while sending message: "+err.Error(), "update", update.UpdateID)
		return
	}
}

func (tg *TelegramClient) sendPrice(update tgbotapi.Update) {
	msg := tgbotapi.NewDocument(update.FromChat().ID, tgbotapi.FilePath("../files/price.pdf"))
	msg.Caption = messages[MsgPrice]
	if _, err := tg.bot.Send(msg); err != nil {
		tg.HandleError("error while sending message: "+err.Error(), "update", update.UpdateID)
		return
	}
}

func (tg *TelegramClient) sendForm(user *types.User, update tgbotapi.Update) {
	msg := tgbotapi.NewMessage(update.FromChat().ID, messages[MsgAskFIO])
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
	re := regexp.MustCompile(`^([A-Za-zА-Яа-яЁё]+[ \t]*)+$`)
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

	msg := tgbotapi.NewMessage(update.FromChat().ID, "Осталось совсем немного 🌝 отправишь свою рабочую / контактную почту?)")
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

	msg := tgbotapi.NewMessage(update.FromChat().ID, messages[MsgAskPhone])
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

	msg := tgbotapi.NewMessage(update.FromChat().ID, messages[MsgAskOrgName])
	msg.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(messages[ButtonNoOrg], "no_org"),
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

		msg := tgbotapi.NewMessage(update.FromChat().ID, messages[MsgAccepted])
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

	msg := tgbotapi.NewMessage(update.FromChat().ID, messages[MsgAskOrgsNumber])
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

	msg := tgbotapi.NewMessage(update.FromChat().ID, messages[MsgAccepted])
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
