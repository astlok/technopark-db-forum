package models

//easyjson:json
type Forum struct {
	ID      uint64 `json:"id,omitempty"`
	Title   string `json:"title,omitempty" db:"title"`
	User    string `json:"user,omitempty" db:"user_nickname"`
	Slug    string `json:"slug,omitempty" db:"slug"`
	Posts   uint64 `json:"posts" db:"posts"`
	Threads uint64 `json:"threads" db:"threads"`
}
