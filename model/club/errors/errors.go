package errors

import "playcards/utils/errors"

var (
	ErrClubNotExisted   = errors.NotFound(14001, "俱乐部不存在！")
	ErrAlreadyInClub    = errors.Conflict(14002, "已有俱乐部，不能重复加入！")
	ErrStatusNoINNormal = errors.NotFound(14003, "俱乐部不在活动状态，不能操作！")
	ErrClubMemberLimit  = errors.NotFound(14004, "该俱乐部已满！")
	ErrInBlackList      = errors.NotFound(14005, "用戶在该俱乐部的黑名单中，不能加入！")
	ErrNotJoinAnyClub   = errors.Conflict(14006, "未加入任何俱乐部！")
	ErrNotInClub        = errors.Conflict(14007, "未加入这个俱乐部！")
	ErrClubNameLen      = errors.Forbidden(14008, "俱乐部名称必须在1到20个汉字或3到60字母之间！")
	ErrCreatorid        = errors.Forbidden(14009, "创建人ID和代理人ID不能为空！")
	ErrExistedInClub    = errors.Conflict(14010, "已加入这个俱乐部！")
	ErrClubRecharge     = errors.Conflict(14011, "充值金额不能为负数！")
	ErrClubConsume      = errors.Conflict(14012, "消费金额不能为正数！")
)
