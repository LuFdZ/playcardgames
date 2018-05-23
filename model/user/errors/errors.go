package errors

import "playcards/utils/errors"

var (
	ErrInvalidToken             = errors.BadRequest(10001, "toekn错误！")
	ErrUserNotExisted           = errors.NotFound(10002, "用户不存在！")
	ErrBillNotExisted           = errors.NotFound(10003, "用户账户信息不存在！")
	ErrInvalidUserInfo          = errors.BadRequest(10004, "非法的用户信息！")
	ErrWXRequest                = errors.BadRequest(10005, "微信服务器请求异常！")
	ErrWXResponse               = errors.BadRequest(10006, "微信服务器请求返回异常！")
	ErrWXResponseJson           = errors.BadRequest(10007, "微信服务器返回解析异常！")
	ErrWXLoginParam             = errors.BadRequest(10008, "OpenID或Code为空！")
	ErrWXParam                  = errors.BadRequest(10009, "微信登录参数错误！")
	ErrParamTooLong             = errors.BadRequest(10010, "参数过长！")
	ErrNameOrPasswordNotExisted = errors.NotFound(10011, "账号名或密码错误！")
	ErrUnionIDNoFind            = errors.BadRequest(10012, "UnionID获取失败！")
	ErrWXLoginResponse          = errors.BadRequest(10014, "微信请求返回错误！%s")
	ErrTest                     = errors.BadRequest(10015, "错误上报测试")
)
