package errors

import "playcards/utils/errors"

var (
	ErrExisted   = errors.NotFound(12001, "请求发送，不能重复操作！")
	ErrNoExisted = errors.Conflict(12002, "对象数据不存在！")
	ErrStatus    = errors.NotFound(12003, "数据状态未改变！")
	ErrUserErr   = errors.NotFound(12004, "玩家不存在！")
)
