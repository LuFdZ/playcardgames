package errors

import "playcards/utils/errors"

var (
	ErrUserNotInGame  = errors.NotFound(15001, "您不在游戏中！")
	ErrNotYouTrun     = errors.Forbidden(15002, "还没轮到您操作！")
	ErrGameNotExist   = errors.NotFound(15003, "游戏信息不存在！")
	ErrGameDiffer     = errors.NotFound(15004, "请求的游戏ID与当前游戏不符！")
	ErrGetBankerParam = errors.NotFound(15005, "抢庄参数不对！")
	ErrGetCardList    = errors.NotFound(15006, "获取牌组异常！")
	ErrCardNotExist   = errors.NotFound(15007, "提交的牌在牌组中不存在！")
)

var (
	ErrBankerDone     = errors.Forbidden(15011, "不是抢庄阶段！")
	ErrSubmitCardDone = errors.Forbidden(15013, "不是出牌阶段！")
	ErrParam          = errors.Conflict(15014, "参数不符合要求！")
	ErrBankerType     = errors.Forbidden(15015, "抢庄类型非法！")
	ErrSubmitCard     = errors.Forbidden(15016, "提交的牌型非法！")
)

var (
	ErrGoLua = errors.Conflict(15017, "go-lua 操作异常！")
)
