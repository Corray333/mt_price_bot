package types

type User struct {
	IsAdmin bool `json:"isAdmin" db:"id_admin"`
}
