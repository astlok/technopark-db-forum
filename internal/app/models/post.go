package models

import (
	"time"
)

type Post struct {
	ID       uint64    `json:"id,omitempty" db:"id"`
	Parent   int       `json:"parent" db:"parent"`
	Author   string    `json:"author,omitempty" db:"author_nickname"`
	Message  string    `json:"message,omitempty" db:"message"`
	IsEdited bool      `json:"isEdited" db:"is_edited"`
	Forum    string    `json:"forum,omitempty" db:"forum_slug"`
	Thread   uint64    `json:"thread,omitempty" db:"thread_id"`
	Tree     string    `json:"-" db:"tree"`
	Created  time.Time `json:"created,omitempty" db:"created"`
}

type PostInfo struct {
	Post   *Post   `json:"post,omitempty"`
	Author *User   `json:"author,omitempty"`
	Thread *Thread `json:"thread,omitempty"`
	Forum  *Forum  `json:"forum,omitempty"`
}