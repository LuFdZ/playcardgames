package errors

import "playcards/utils/errors"

var (
	ErrUserNotExisted     = errors.NotFound(10002, "user not existed")
	ErrPropertyNotExisted = errors.NotFound(10003, "property not existed")
	ErrInvalidRoomInfo    = errors.BadRequest(10004, "invalid user info")
)
