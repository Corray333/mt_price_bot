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
	GetUserByID(user_id int64) (*types.User, error)
	GetAllAdmins() ([]*types.User, error)
}

const (
	MsgWelcome = iota + 1
	MsgAskFIO
	MsgAskEmail
	MsgAskPhone
	MsgAskOrgName
	MsgAskOrgsNumber
	MsgPrice
	MsgError
	MsgAccepted
	ButtonPrice
	ButtonForm
	ButtonNoOrg
)

type TelegramClient struct {
	bot   *tgbotapi.BotAPI
	store Storage
}

var messages = map[int]string{
	MsgWelcome:       "Привет, в этом боте ты можешь запросить актуальный прайс и оставить заявку",
	MsgAskFIO:        "Чтобы оставить заявку, отправьте свои ФИО",
	MsgAskEmail:      "Теперь отправьте свою почту",
	MsgAskPhone:      "Отправьте контактный номер телефона",
	MsgAskOrgName:    "Отправьте название вашей точки (если есть)",
	MsgAskOrgsNumber: "Сколько у вас точек?",
	MsgPrice:         "Наш актуальный прайс",
	MsgError:         "Что-то пошло не так, попробуйте снова",
	MsgAccepted:      "Спасибо, ваша заявка принята",
	ButtonPrice:      "Получить прайс",
	ButtonForm:       "Оставить заявку",
	ButtonNoOrg:      "Нет точки",
}

var admins = []string{
	"corray9",
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

		user, err := tg.store.GetUserByID(update.FromChat().ID)
		if err != nil {
			if err == sql.ErrNoRows {
				tg.sendWelcomeMessage(update)
				continue
			}
			tg.HandleError("error while getting user from db: "+err.Error(), "update_id", update.UpdateID)
			continue
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

	if update.Message != nil {
		if update.Message.Text == messages[ButtonPrice] {
			tg.sendPrice(update)
			return
		} else if update.Message.Text == messages[ButtonForm] {
			tg.sendForm(user, update)
			return
		}
	}

	switch user.State {
	case StateWaitingFIO:
		tg.handleInputFIO(user, update)
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

func (tg *TelegramClient) handleAdminUpdate(user *types.User, update tgbotapi.Update) {

}
