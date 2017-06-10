package errors

import "playcards/utils/errors"

var (
	ErrNoPermission = errors.Forbidden(13001, "no permission")
	ErrNoMethod     = errors.NotFound(13002, "no method")
	ErrNoToken      = errors.Unauthorized(13003, "no token")
	ErrUnauthorized = errors.Unauthorized(13004, "unauthorized")
	ErrInvalidToken = errors.Unauthorized(13005, "token invalid")
)
