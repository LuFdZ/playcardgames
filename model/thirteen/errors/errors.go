package errors

import "playcards/utils/errors"

var (
	ErrUserNotInGame = errors.Conflict(50001, "您不在游戏中！")
	ErrUserAlready   = errors.NotFound(50002, "您已经提交过牌组了！")
)
