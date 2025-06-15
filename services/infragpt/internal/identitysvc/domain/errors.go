package domain

import "errors"

var (
	ErrDuplicateKey = errors.New("duplicate key constraint violation")
)