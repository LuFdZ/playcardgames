package errors

import "playcards/utils/errors"

const (
	ErrDiamondNotEnough = 40000
)

var (
	ErrRoomNotExisted    = errors.NotFound(40001, "房间不存在！")
	ErrInvalidRoomInfo   = errors.Conflict(40002, "用户信息错误！")
	ErrUserNotInRoom     = errors.NotFound(40003, "您不在房间中！")
	ErrUserAlreadyInRoom = errors.Forbidden(40004, "您已经在房间中！")
	ErrRoomFull          = errors.Conflict(40005, "房间已满！")
	ErrGameHasBegin      = errors.Forbidden(40006, "房间正在游戏中！")
	ErrNotReadyStatus    = errors.Forbidden(40007, "不能操作，房间不在准备状态！")
	ErrRoomPwdExisted    = errors.Conflict(40008, "此房间编码已经存在！")
	ErrNotEnoughDiamond  = errors.Conflict(40009, "钻石不足！")
	ErrGameIsDone        = errors.Forbidden(40010, "游戏已经结束！")
	ErrInGiveUp          = errors.Forbidden(40011, "正在进行放弃游戏投票！")
	ErrAlreadyVoted      = errors.Forbidden(40012, "您已经投过票了！")
	ErrNotInSameRoon     = errors.Forbidden(40013, "已请求对象不在同一房间内！")
	ErrRenewalRoon       = errors.Forbidden(40014, "房间已续费或已结束！")
	ErrAlreadyReady      = errors.Forbidden(40015, "您已经准备过了！")
	ErrNotInGiveUp       = errors.Forbidden(40016, "游戏不在放弃投票状态！")
	ErrRoomBusy          = errors.Forbidden(40017, "房间繁忙，请稍后再试！")
	ErrRoomNotFind       = errors.NotFound(40018, "房间不存在！")
	ErrRoomAgentLimit    = errors.NotFound(40019, "代开房间数量已达到上限！")
	ErrNotPayer          = errors.Forbidden(40020, "不是您的代开房间不能解散！")
	ErrRoomType          = errors.Forbidden(40021, "本房间不能续费操作！")
	ErrNotClubMember     = errors.Forbidden(40022, "不是俱乐部成员！")
	ErrClubCantRenewal   = errors.Forbidden(40023, "俱乐部房间不能续费！")
)
