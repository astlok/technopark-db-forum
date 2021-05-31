package models

//easyjson:json
type NumRecords struct {
	User   uint64 `json:"user" db:"user_count"`
	Forum  uint64 `json:"forum" db:"forum_count"`
	Thread uint64 `json:"thread" db:"thread_count"`
	Post   uint64 `json:"post" db:"post_count"`
}
