package errors

import "playcards/utils/errors"

var (
	ErrUserNotInGame     = errors.Conflict(16001, "您不在游戏中！")
	ErrAlreadySetBet     = errors.NotFound(16003, "您已经提交过下注请求了！")
	ErrAlreadySubmitCard = errors.NotFound(16004, "您已经提交过牌组了！")
	ErrGameNotExist      = errors.NotFound(16005, "游戏信息不存在！")
	ErrBankerNoBet       = errors.NotFound(16006, "庄家不能下注！")
	ErrGameNoFind        = errors.Conflict(16007, "未找到游戏信息！")
	ErrBetDone           = errors.Conflict(16012, "不是下注阶段！")
	ErrSubmitCardDone    = errors.Conflict(16013, "不是开牌阶段！")
	ErrParam             = errors.Conflict(16014, "参数不符合要求！")
	ErrCardNotExist      = errors.NotFound(16016, "提交的牌在牌组中不存在！")
	ErrGoLua             = errors.Conflict(16017, "go-lua 操作异常！")
)