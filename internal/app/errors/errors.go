package errors

import "errors"

var (
	ErrDuplicate      = errors.New("duplicate")
	ErrNoParent       = errors.New("any post has no parent")
	ErrConflict       = errors.New("user conflict")
	ErrUserNotFound   = errors.New("user not found")
	ErrForumNotFound  = errors.New("forum not found")
	ErrThreadNotFound = errors.New("thread not found")
	ErrPostNotFound   = errors.New("post not found")
)
