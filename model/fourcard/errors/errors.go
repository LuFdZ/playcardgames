package errors

import "playcards/utils/errors"

var (
	ErrUserNotInGame     = errors.Conflict(60001, "您不在游戏中！")
	ErrAlreadyGetBanker  = errors.NotFound(60002, "您已经提交过抢庄请求了！")
	ErrAlreadySetBet     = errors.NotFound(60003, "您已经提交过下注请求了！")
	ErrAlreadySubmitCard = errors.NotFound(60004, "您已经提交过牌组了！")
	ErrGameNotExist      = errors.NotFound(60005, "游戏信息不存在！")
	ErrBankerNoBet       = errors.NotFound(60006, "庄家不能下注！")
	ErrGameNoFind        = errors.Conflict(60007, "未找到游戏信息！")
)

var (
	ErrDiceDone       = errors.Conflict(60011, "不是定序阶段！")
	ErrBetDone        = errors.Conflict(60012, "不是下注阶段！")
	ErrSubmitCardDone = errors.Conflict(60013, "不是开牌阶段！")
	ErrParam          = errors.Conflict(60014, "参数不符合要求！")
	ErrCardNotExist   = errors.NotFound(60015, "提交的牌在牌组中不存在！")
)

var (
	ErrGoLua = errors.Conflict(60015, "go-lua 操作异常！")
)
