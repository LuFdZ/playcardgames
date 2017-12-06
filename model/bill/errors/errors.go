package errors

import "playcards/utils/errors"

var (
	ErrInvalidParameter = errors.Forbidden(20001, "invalid parameter")
	ErrOutOfBalance     = errors.Forbidden(20002, "金额不足！")
	ErrNotAllowAmount   = errors.Forbidden(20003, "金额不能为零！")
	ErrFreezeAmount     = errors.Forbidden(20004, "冻结金额总数不能为负！")
)
