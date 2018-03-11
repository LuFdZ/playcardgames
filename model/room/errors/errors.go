package errors

import (
	"playcards/utils/errors"
)

const (
	ErrDiamondNotEnough = 41000
	ErrRoomDisband      = 41028
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
	ErrRenewalRoon       = errors.Forbidden(40014, "房间已续费！")
	ErrAlreadyReady      = errors.Forbidden(40015, "您已经准备过了！")
	ErrNotInGiveUp       = errors.Forbidden(40016, "游戏不在放弃投票状态！")
	//ErrRoomBusy          = errors.Forbidden(40017, "房间繁忙，请稍后再试！")
	ErrRoomNotAllReady         = errors.NotFound(40018, "有玩家未准备好！")
	ErrRoomAgentLimit          = errors.NotFound(40019, "代开房间数量已达到上限！")
	ErrNotPayer                = errors.Forbidden(40020, "您无权操作此房间！")
	ErrRoomType                = errors.Forbidden(40021, "本房间不能续费操作！")
	ErrNotClubMember           = errors.Forbidden(40022, "不是俱乐部成员！")
	ErrClubCantRenewal         = errors.Forbidden(40023, "俱乐部房间不能续费！")
	ErrGameParam               = errors.Forbidden(40024, "房间参数不正确！")
	ErrRoomMaxNumber           = errors.Forbidden(40025, "房间人数不正确！")
	ErrRoomMaxRound            = errors.Forbidden(40026, "房间轮数不正确！")
	ErrPlayerNumberNoEnough    = errors.NotFound(40027, "玩家人数不足！")
	ErrGameType                = errors.NotFound(40028, "游戏类型不允许该操作！")
	ErrShuffle                 = errors.Forbidden(40029, "您不能洗牌！")
	ErrClubCreaterNotFind      = errors.Forbidden(40030, "俱乐部创建人信息未找到！")
	ErrNotEnoughGold           = errors.Conflict(40031, "金币不足！")
	ErrAlreadyInBankerList     = errors.Conflict(40032, "您已经在申请上庄列表中！")
	ErrNoInBankerList          = errors.Conflict(40033, "您不在在申请上庄列表中！")
	ErrOutBankerList           = errors.Conflict(40034, "现在不能下庄！")
	ErrNotEnoughClubCoin       = errors.Conflict(40035, "俱乐部奖杯不足！最少需要%d")
	ErrOutBankerListWithBanker = errors.Conflict(40036, "你是当前你庄家，此轮游戏结束前不能下庄！")
	ErrSettingParam            = errors.Forbidden(40037, "比赛模式下，设置不能为空！")
	ErrCanNotIntoClubRoom       = errors.Conflict(40038, "您不在这个俱乐部或被禁，详情请联系俱乐部会长！")
)
