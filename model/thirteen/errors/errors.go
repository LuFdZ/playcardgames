package errors

import "playcards/utils/errors"

var (
	ErrUserNotInGame = errors.Conflict(50001, "you are not in the game")
	ErrUserAlready   = errors.NotFound(50002, "you have already submitted cards")
)
