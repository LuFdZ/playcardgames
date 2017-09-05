package enum

const (
	ErrUID = -1
)

const (
	LoopTime       = 3
	MaxRecordCount = 50
)

const (
	RoomTableName = "rooms"
)

const (
	RoomStatusInit        = 1 //建立房间
	RoomStatusAllReady    = 2 //房间满员并且全部确认准备
	RoomStatusStarted     = 3 //游戏开始
	RoomStatusReInit      = 4 //一局游戏结束，等待下一轮确认开始
	RoomStatusDelay       = 5 //房间到达游戏最大局数，留存一段时间等待同房续费
	RoomStatusDone        = 6 //房间正常结束
	RoomStatusDestroy     = 7 //房间非正常结束(所有人员离开或者弃权)
	RoomStatusWairtGiveUp = 8 //发起投票弃权的等待状态
)

const (
	UserRoleMaster = 1
	UserRoleSlave  = 2
)

const (
	UserUnready = 0
	UserReady   = 1
)

const (
	ThirteenGameCost = 1
)

const (
	GiveupStatusAgree    = 1
	GiveupStatusDisAgree = 2
	GiveupStatusWairting = 3
)
