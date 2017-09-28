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
	RoomStatusInit          = 1  //建立房间
	RoomStatusAllReady      = 2  //房间满员并且全部确认准备
	RoomStatusStarted       = 3  //游戏开始
	RoomStatusReInit        = 4  //一局游戏结束，等待下一轮确认开始
	RoomStatusWaitGiveUp    = 5  //发起投票弃权的等待状态
	RoomStatusDelay         = 6  //房间到达游戏最大局数，留存一段时间等待同房续费
	RoomStatusDone          = 7  //房间正常结束
	RoomStatusDestroy       = 8  //房间非正常结束(所有人员离开)
	RoomStatusGiveUp        = 9  //房间投票放弃
	RoomStatusOverTimeClean = 10 //房间长时间无人操作被超时清除
)

// room user role
const (
	UserRoleMaster = 1
	UserRoleSlave  = 2
)

// room user ready status
const (
	UserReady   = 1
	UserUnready = 2
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

// web socket conncthion status
const (
	SocketAline = 1
	SocketClose = 2
)

const (
	UserStateLeader  = iota + 1 //带头解散的人
	UserStateOppose             //已反对
	UserStateAgree              //已同意
	UserStateOffline            //离线
	UserStateWaiting            //等待操作
)

const (
	AgreeGiveUpRoom    = 1
	DisAgreeGiveUpRoom = 2
)

const (
	ThirteenGameType = 1001
	NiuniuGameType   = 1002
)
