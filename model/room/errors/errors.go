package errors

import "playcards/utils/errors"

var (
	ErrRoomNotExisted    = errors.NotFound(40001, "room not existed")
	ErrInvalidRoomInfo   = errors.BadRequest(40003, "invalid user info")
	ErrUserNotInRoom     = errors.Conflict(40004, "you are not in the room")
	ErrUserAlreadyInRoom = errors.Conflict(40005, "you are already in the room")
	ErrRoomFull          = errors.Conflict(40006, "the room is full")
	ErrGameHasBegin      = errors.Conflict(40007, "The game is already begining")
	ErrNotReadyStatus    = errors.Conflict(40008, "Not in the ready state")
	ErrRoomPwdExisted    = errors.Conflict(40009, "plase change password")
	ErrNotEnoughDiamond  = errors.Conflict(40010, "diamond not enough")
	ErrGameIsDone        = errors.Conflict(40011, "The game is done")
)
