package errors

import "playcards/utils/errors"

var (
	ErrMailInfoNotFind    = errors.NotFound(21001, "邮件配置信息未找到！")
	ErrMailSendLogNotFind = errors.NotFound(21002, "邮件发送记录未找到！")
	ErrMailInfoID         = errors.NotFound(21003, "模板ID缺失！")
	ErrMailInfoContent    = errors.NotFound(21004, "邮件内容缺失！")
	ErrMailNotFind        = errors.NotFound(21005, "邮件已过期！")
	ErrArealyGetMailItem  = errors.NotFound(21006, "邮件已领取过！")
	ErrHasNotItem         = errors.NotFound(21007, "该邮件没有附件！")
	ErrItemFormat         = errors.NotFound(21008, "邮件附件格式错误！")
	ErrSendAndChannel     = errors.NotFound(21009, "发送人和发送渠道不能同时为空！")
	ErrMailNotNow         = errors.NotFound(21010, "沒有新邮件！")
)
