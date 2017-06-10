package errors

import "playcards/utils/errors"

var (
	ErrInvalidParameter = errors.Forbidden(20001, "invalid parameter")
	ErrOutOfBalance     = errors.Forbidden(20002, "out of balance")
	ErrNotAllowAmount   = errors.Forbidden(20003, "not allow amount")
)
