package models

//easyjson:json
type UserList []User

//easyjson:json
type User struct {
	Nickname string `json:"nickname,omitempty" db:"nickname"`
	Fullname string `json:"fullname,omitempty" db:"fullname"`
	About    string `json:"about,omitempty" db:"about"`
	Email    string `json:"email,omitempty" db:"email"`
}
