package errors

import "playcards/utils/errors"

var (
	ErrRoomNotExisted = errors.NotFound(40001, "The room does not exist!")
	//房间不存在！
	ErrInvalidRoomInfo = errors.BadRequest(40003, "User information error!")
	//用户信息错误！
	ErrUserNotInRoom = errors.Conflict(40004, "You are not in the room!")
	//您不在房间中！
	ErrUserAlreadyInRoom = errors.Conflict(40005, "You are already in the room!")
	//您已经在房间中！
	ErrRoomFull = errors.Conflict(40006, "The room is full!")
	//房间已满！
	ErrGameHasBegin = errors.Conflict(40007, "The game has already started!")
	//游戏已经开始！
	ErrNotReadyStatus = errors.Conflict(40008, "The room is not ready!")
	//不能操作，房间不在准备状态！
	ErrRoomPwdExisted = errors.Conflict(40009, "This room code already exists!")
	//此房间编码已经存在！
	ErrNotEnoughDiamond = errors.Conflict(40010, "Diamond not enguth")
	//钻石不足！
	ErrGameIsDone = errors.Conflict(40011, "The game is over!")
	//游戏已经结束！
	ErrInGiveUp = errors.Conflict(40012, "Dropping a game vote!")
	//正在进行放弃游戏投票！
	ErrAlreadyVoted = errors.Conflict(40013, "You have already voted!")
	//您已经投过票了！
	ErrNotInSameRoon = errors.Conflict(40014, "Not in the same room")
	//已请求对象不在同一房间内
)
