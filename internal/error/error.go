package errconst

import "errors"

var (
	ErrInvalidURL = errors.New("invalid url")
	ErrNotFound   = errors.New("url not found")
	ErrAliasTaken = errors.New("alias already taken")
)
