package errors

import "playcards/utils/errors"

var (
	ErrRoomNotExisted    = errors.NotFound(40001, "房间不存在！")
	ErrInvalidRoomInfo   = errors.Conflict(40002, "用户信息错误！")
	ErrUserNotInRoom     = errors.NotFound(40003, "您不在房间中！")
	ErrUserAlreadyInRoom = errors.Forbidden(40004, "您已经在房间中！")
	ErrRoomFull          = errors.Conflict(40005, "房间已满！")
	ErrGameHasBegin      = errors.Forbidden(40006, "游戏已经开始！")
	ErrNotReadyStatus    = errors.Forbidden(40007, "不能操作，房间不在准备状态！")
	ErrRoomPwdExisted    = errors.Conflict(40008, "此房间编码已经存在！")
	ErrNotEnoughDiamond  = errors.Conflict(40009, "钻石不足！")
	ErrGameIsDone        = errors.Forbidden(40010, "游戏已经结束！")
	ErrInGiveUp          = errors.Forbidden(40011, "正在进行放弃游戏投票！")
	ErrAlreadyVoted      = errors.Forbidden(40012, "您已经投过票了！")
	ErrNotInSameRoon     = errors.Forbidden(40013, "已请求对象不在同一房间内！")
	ErrRenewalRoon       = errors.Forbidden(40014, "房间游戏未结束或非正常结束，不能续费！")
	ErrAlreadyReady      = errors.Forbidden(40015, "您已经准备过了！")
)
