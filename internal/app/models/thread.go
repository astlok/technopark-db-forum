package models

import (
	"time"
)

type Thread struct {
	ID      uint64    `json:"id,omitempty" db:"id"`
	Title   string    `json:"title,omitempty" db:"title"`
	Author  string    `json:"author,omitempty" db:"author_nickname"`
	Forum   string    `json:"forum,omitempty" db:"forum_slug"`
	Message string    `json:"message,omitempty" db:"message"`
	Votes   int       `json:"votes" db:"votes"`
	Slug    string    `json:"slug,omitempty" db:"slug"`
	Created time.Time `json:"created,omitempty" db:"created"`
}

type Vote struct {
	Nickname string `json:"nickname,omitempty" db:"nickname"`
	Voice    int    `json:"voice,omitempty" db:"voice"`
}
