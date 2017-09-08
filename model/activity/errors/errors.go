package errors

import "playcards/utils/errors"

var (
	ErrActivityNotExisted             = errors.NotFound(90001, "delete activity not found")
	ErrDayStartGreaterDayEnd          = errors.Conflict(91001, "day start greater than day end")
	ErrDurationTooLarge               = errors.Conflict(91002, "duration too large")
	ErrTimeStartDayNotEqualTimeEndDay = errors.Conflict(91003, "time day start not equal time day end")
	ErrHadInviter                     = errors.Conflict(91004, "已填写过邀请人！")
	ErrInviterNotNewUser              = errors.Conflict(91005, "您不是新用户！")
)
