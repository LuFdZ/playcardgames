package errors

import "playcards/utils/errors"

var (
	ErrMailInfoNotFind    = errors.NotFound(21001, "邮件配置信息未找到！")
	ErrMailSendLogNotFind = errors.NotFound(21002, "邮件发送记录未找到！")
	ErrMailInfoID         = errors.NotFound(21003, "模板ID缺失！")
	ErrMailInfoContent    = errors.NotFound(21004, "邮件内容缺失！")
)
