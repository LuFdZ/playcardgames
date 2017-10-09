package errors

import "playcards/utils/errors"

var (
	ErrInvalidToken       = errors.BadRequest(10001, "toekn错误！")
	ErrUserNotExisted     = errors.NotFound(10002, "用户不存在！")
	ErrPropertyNotExisted = errors.NotFound(10003, "用户属性不存在！")
	ErrInvalidUserInfo    = errors.BadRequest(10004, "非法的用户信息！")
	ErrWXRequest          = errors.BadRequest(10005, "微信服务器请求异常！")
	ErrWXResponse         = errors.BadRequest(10006, "微信服务器请求返回异常！")
	ErrWXResponseJson     = errors.BadRequest(10007, "微信服务器返回解析异常！")
	ErrWXLoginParam       = errors.BadRequest(10008, "OpenID或Code为空！")
	ErrWXParam            = errors.BadRequest(10009, "微信登录参数错误！")
)
