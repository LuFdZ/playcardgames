package errors

import "playcards/utils/errors"

var (
	ErrUserNotInGame     = errors.Conflict(23001, "您不在游戏中！")
	ErrAlreadySetBet     = errors.NotFound(23003, "您已经提交过下注请求了！")
	ErrAlreadySubmitCard = errors.NotFound(23004, "您已经提交过牌组了！")
	ErrGameNotExist      = errors.NotFound(23005, "游戏信息不存在！")
	ErrBankerNoBet       = errors.NotFound(23006, "庄家不能下注！")
	ErrGameNoFind        = errors.Conflict(23007, "未找到游戏信息！")
	ErrBetDone           = errors.Conflict(23012, "不是下注阶段！")
	ErrSubmitCardDone    = errors.Conflict(23013, "不是开牌阶段！")
	ErrParam             = errors.Conflict(23014, "参数不符合要求！")
	//ErrCardNotExist      = errors.NotFound(23016, "提交的牌在牌组中不存在！")
	ErrGoLua             = errors.Conflict(23017, "go-lua 操作异常！")
)