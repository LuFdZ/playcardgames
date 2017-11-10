package errors
import "playcards/utils/errors"

var (
	ErrActivityNotExisted             = errors.NotFound(90001, "活动信息不存在！")
	ErrDayStartGreaterDayEnd          = errors.Conflict(91002, "活动的开始时间不能大于结束时间！")
	ErrDurationTooLarge               = errors.Forbidden(91003, "持续时间过长！")
	ErrTimeStartDayNotEqualTimeEndDay = errors.Conflict(91004, "开始时间不能等于结束时间！")
	ErrIDNotFIND                      = errors.NotFound(90005, "未找到数据ID！")
)

var (
	ErrHadInviter      = errors.Forbidden(91005, "已填写过邀请人！")
	ErrInviterSelf     = errors.Forbidden(91006, "不能邀请自己！")
	ErrInviterNoExist  = errors.Forbidden(91007, "无效的推荐人ID，请重新输入！")
	ErrInviterConflict = errors.Forbidden(91008, "您已邀请过该用户，该用户不能作为您的邀请人！")
	ErrInviterOverdue  = errors.Forbidden(91009, "填写成功，但您注册已超过3天，无法获取奖励水晶！")
)

var (
	ErrShareNoDiamonds = errors.Conflict(91010, "不符合分享获取水晶条件！")
)
