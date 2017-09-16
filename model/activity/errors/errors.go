package errors

import "playcards/utils/errors"

var (
	ErrActivityNotExisted             = errors.NotFound(90001, "活动信息不存在！")
	ErrDayStartGreaterDayEnd          = errors.Conflict(91002, "活动的开始时间不能大于结束时间！")
	ErrDurationTooLarge               = errors.Forbidden(91003, "持续时间过长！")
	ErrTimeStartDayNotEqualTimeEndDay = errors.Conflict(91004, "开始时间不能等于结束时间！")
)

var (
	ErrHadInviter     = errors.Forbidden(91005, "已填写过邀请人！")
	ErrInviterSelf    = errors.Forbidden(91006, "不能邀请自己！")
	ErrInviterNoExist = errors.Forbidden(91007, "无效的推荐人ID，请重新输入！")
)

var (
	ErrShareNoDiamonds = errors.Conflict(91008, "不符合分享获取水晶条件！")
)
