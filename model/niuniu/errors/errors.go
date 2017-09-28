package errors

import "playcards/utils/errors"

var (
	ErrUserNotInGame     = errors.Conflict(60001, "您不在游戏中！")
	ErrAlreadyGetBanker  = errors.NotFound(60002, "您已经提交过抢庄请求了！")
	ErrAlreadySetBet     = errors.NotFound(60003, "您已经提交过下注请求了！")
	ErrAlreadySubmitCard = errors.NotFound(60004, "您已经提交过牌组了！")
	ErrGameNotExist      = errors.NotFound(60005, "游戏信息不存在！")
)

var (
	ErrBankerDone     = errors.Conflict(60011, "抢庄阶段已过！")
	ErrBetDone        = errors.Conflict(60012, "下注阶段已过！")
	ErrSubmitCardDone = errors.Conflict(60013, "开牌阶段已过！")
	ErrParam          = errors.Conflict(60014, "参数不符合要求！")
)
