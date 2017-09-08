package enum

//error config
const (
	ErrUID = -1
)

//globle config
const (
	LoopTime         = 3
	MaxRecordCount   = 50
	RoomDelayMinutes = 5.0
	RoomCodeMax      = 999999
	RoomCodeMin      = 100000
)

//db config
const (
	RoomTableName       = "rooms"
	PlayerRoomTableName = "player_rooms"
)

// room status desc
const (
	RoomStatusInit        = 1 //建立房间
	RoomStatusAllReady    = 2 //房间满员并且全部确认准备
	RoomStatusStarted     = 3 //游戏开始
	RoomStatusReInit      = 4 //一局游戏结束，等待下一轮确认开始
	RoomStatusDelay       = 5 //房间到达游戏最大局数，留存一段时间等待同房续费
	RoomStatusDone        = 6 //房间正常结束
	RoomStatusDestroy     = 7 //房间非正常结束(所有人员离开)
	RoomStatusGiveUp      = 8 //房间投票放弃
	RoomStatusWairtGiveUp = 9 //发起投票弃权的等待状态
)

// room user role
const (
	UserRoleMaster = 1
	UserRoleSlave  = 2
)

// room user ready status
const (
	UserUnready = 0
	UserReady   = 1
)

// game cost
const (
	ThirteenGameCost = 1.0
)

// give up vote status
const (
	GiveupStatusAgree    = 1
	GiveupStatusDisAgree = 2
	GiveupStatusWairting = 3
)

// room type desc
const (
	RoomTypeNom   = 1
	RoomTypeAgent = 2
)
