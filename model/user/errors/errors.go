package errors

import "playcards/utils/errors"

var (
	ErrInvalidToken       = errors.BadRequest(10001, "invalid token")
	ErrUserNotExisted     = errors.NotFound(10002, "user not existed")
	ErrPropertyNotExisted = errors.NotFound(10003, "property not existed")
	ErrInvalidUserInfo    = errors.BadRequest(10004, "invalid user info")
)
