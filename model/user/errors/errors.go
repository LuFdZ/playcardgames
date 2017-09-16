package errors

import "playcards/utils/errors"

var (
	ErrInvalidToken       = errors.BadRequest(10001, "toekn错误！")
	ErrUserNotExisted     = errors.NotFound(10002, "用户不存在！")
	ErrPropertyNotExisted = errors.NotFound(10003, "用户属性不存在！")
	ErrInvalidUserInfo    = errors.BadRequest(10004, "非法的用户信息！")
)
