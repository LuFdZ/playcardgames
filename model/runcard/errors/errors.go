package errors

import "playcards/utils/errors"

var (
	ErrUserNotInGame           = errors.Conflict(24001, "您不在游戏中！")
	ErrAlreadySubmitCard       = errors.NotFound(24002, "您已经提交过牌组了！")
	ErrGameNotExist            = errors.NotFound(24003, "游戏信息不存在！")
	ErrGameNoFind              = errors.Conflict(24004, "未找到游戏信息！")
	ErrSubmitCardDone          = errors.Conflict(24005, "不是开牌阶段！")
	ErrGoLua                   = errors.Conflict(24006, "go-lua 操作异常！")
	ErrCardNotExist            = errors.NotFound(24007, "提交的牌在牌组中不存在！")
	ErrNotYourTurn             = errors.Conflict(24008, "还没轮到您操作！")
	ErrSubmitCardValueTooSmall = errors.Conflict(24009, "提交的牌值太小！")
	ErrSubmitCardNil           = errors.Conflict(24010, "不能提交空牌组！")
	ErrSubmitCardType          = errors.Conflict(24011, "牌型非法！")
	ErrHaveToSubmitCard        = errors.Conflict(24012, "有必管模式下必须出牌！")
	ErrSubmitCardFail          = errors.Conflict(24013, "提交失败！")
	ErrStartCardSubmit          = errors.Conflict(24014, "首出必须带上首发牌！")
)
