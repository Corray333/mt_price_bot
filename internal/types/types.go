package types

type User struct {
	ID        int64  `json:"id" db:"user_id"`
	Username  string `json:"username" db:"username"`
	State     int    `json:"state" db:"state"`
	IsAdmin   bool   `json:"isAdmin" db:"is_admin"`
	FIO       string `json:"fio" db:"fio"`
	Phone     string `json:"phone" db:"phone"`
	Email     string `json:"email" db:"email"`
	Org       string `json:"org" db:"org"`
	OrgNumber int    `json:"orgNumber" db:"org_number"`
}
