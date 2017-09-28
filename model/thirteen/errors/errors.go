package errors

import "playcards/utils/errors"

var (
	ErrUserNotInGame     = errors.Conflict(50001, "您不在游戏中！")
	ErrAlreadySubmitCard = errors.NotFound(50002, "您已经提交过牌组了！")
	ErrGameNotExist      = errors.NotFound(50003, "游戏信息不存在！")
	ErrCardNotExist      = errors.NotFound(50004, "提交的牌在牌组中不存在！")
)
