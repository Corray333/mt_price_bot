package storage

import (
	"fmt"
	"os"

	"github.com/Corray333/mt_price_bot/internal/types"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

type Storage struct {
	db *sqlx.DB
}

const (
	MsgWelcome = iota
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

var Messages = map[int]string{
	MsgWelcome:       "В этом боте вы можете запросить актуальный прайс и оставить заявку",
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

var Admins = []string{
	"corray9",
}

func New() *Storage {
	db, err := sqlx.Open("postgres", os.Getenv("DB_CONN_STR"))
	if err != nil {
		panic(err)
	}

	if err := db.Ping(); err != nil {
		panic(err)
	}

	return &Storage{
		db: db,
	}
}

func (s *Storage) UpdateUser(user *types.User) error {
	_, err := s.db.Exec(`UPDATE users SET username = $1, phone = $2, email = $3, org = $4, org_number = $5, fio = $6, state = $7 WHERE user_id = $8`, user.Username, user.Phone, user.Email, user.Org, user.OrgNumber, user.FIO, user.State, user.ID)
	return err

}
func (s *Storage) CreateUser(user *types.User) error {
	fmt.Printf("%+v\n", *user)
	_, err := s.db.Exec(`INSERT INTO users (user_id, username, is_admin) VALUES ($1, $2, $3)`, user.ID, user.Username, user.IsAdmin)
	return err
}

func (s *Storage) GetUserByID(user_id int64) (*types.User, error) {
	user := &types.User{}
	err := s.db.Get(user, `SELECT * FROM users WHERE user_id = $1`, user_id)
	return user, err
}

func (s *Storage) GetAllAdmins() ([]*types.User, error) {
	users := []*types.User{}
	err := s.db.Select(&users, `SELECT * FROM users WHERE is_admin = true`)
	return users, err
}
