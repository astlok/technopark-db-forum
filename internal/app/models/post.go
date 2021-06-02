package models

import (
	"github.com/go-openapi/strfmt"
	"github.com/lib/pq"
)

//easyjson:json
type PostList []Post

//easyjson:json
type Post struct {
	ID       uint64          `json:"id,omitempty" db:"id"`
	Parent   int             `json:"parent" db:"parent"`
	Author   string          `json:"author,omitempty" db:"author_nickname"`
	Message  string          `json:"message,omitempty" db:"message"`
	IsEdited bool            `json:"isEdited" db:"is_edited"`
	Forum    string          `json:"forum,omitempty" db:"forum_slug"`
	Thread   uint64          `json:"thread,omitempty" db:"thread_id"`
	Tree     pq.Int64Array   `json:"-" db:"tree"`
	Created  strfmt.DateTime `json:"created,omitempty" db:"created"`
}

//easyjson:json
type PostInfo struct {
	Post   *Post   `json:"post,omitempty"`
	Author *User   `json:"author,omitempty"`
	Thread *Thread `json:"thread,omitempty"`
	Forum  *Forum  `json:"forum,omitempty"`
}
