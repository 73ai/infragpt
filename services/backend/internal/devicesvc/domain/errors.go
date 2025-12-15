package domain

import "errors"

var (
	ErrDeviceCodeNotFound  = errors.New("device code not found")
	ErrDeviceCodeExpired   = errors.New("device code expired")
	ErrDeviceCodeUsed      = errors.New("device code already used")
	ErrDeviceTokenNotFound = errors.New("device token not found")
	ErrDeviceTokenRevoked  = errors.New("device token revoked")
	ErrDeviceTokenExpired  = errors.New("device token expired")
	ErrInvalidUserCode     = errors.New("invalid user code")
	ErrAuthorizationPending = errors.New("authorization pending")
)
