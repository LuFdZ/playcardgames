package errors

import "playcards/utils/errors"

var (
	ErrRoomNotExisted     = errors.NotFound(40001, "room not existed")
	ErrPropertyNotExisted = errors.NotFound(40002, "property not existed")
	ErrInvalidRoomInfo    = errors.BadRequest(40003, "invalid user info")
	ErrNotInRoom          = errors.Conflict(40004, "you are not in the room")
	ErrRoomAlreadyInRoom  = errors.Conflict(40005, "you are already in the room")
	ErrRoomFull           = errors.Conflict(40006, "the room is full")
	ErrGameHasBegin       = errors.Conflict(40007, "The game is already begining")
	ErrNotReadyStatus     = errors.Conflict(40008, "Not in the ready state")
)
