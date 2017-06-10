package errors

import "playcards/utils/errors"

var (
	ErrInvalidTimeRange = errors.BadRequest(11001, "invalid timerange")
)
