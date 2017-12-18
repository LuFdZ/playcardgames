package enum

//error config
const (
	ErrUID         = -1
	AgentRoomLimit = 10
	LogTime        = 20
)

//globle config
const (
	LoopTime               = 500
	MaxRecordCount         = 50
	RoomDelayMinutes       = 15.0
	RoomCodeMax            = 999999
	RoomCodeMin            = 100000
	RoomGiveupCleanMinutes = 5.0
)

//db config
const (
	RoomTableName       = "rooms"
	PlayerRoomTableName = "player_rooms"
)

// room status desc
const (
	RoomStatusInit            = 1  //建立房间
	RoomStatusAllReady        = 2  //房间满员并且全部确认准备
	RoomStatusStarted         = 3  //游戏开始
	RoomStatusReInit          = 4  //一局游戏结束，等待下一轮确认开始
	RoomStatusDelay           = 5  //房间到达游戏最大局数，留存一段时间等待同房续费
	RoomStatusDone            = 6  //房间正常结束
	RoomStatusDestroy         = 7  //房间非正常结束(所有人员离开)
	RoomStatusGiveUp          = 8  //房间投票放弃
	RoomStatusOverTimeClean   = 9  //房间长时间无人操作被超时清除
	RoomStatusDiamondNoEnough = 10 //房间开始时付费人水晶不足 取消房间数据
)

// room type desc
const (
	RoomTypeNom   = 1 //普通房
	RoomTypeAgent = 2 //代理开房
	RoomTypeClub  = 3 //俱乐部
)

// give up vote status
const (
	GiveupStatusAgree    = 1 //同意投降
	GiveupStatusDisAgree = 2 //不同意投降
	GiveupStatusWairting = 3 //继续等待投票
)

const (
	NoGiveUp   = 1 //游戏不在投降状态
	WaitGiveUp = 2 //游戏进入投降状态
)

const (
	UserStateLeader   = iota + 1 //带头解散的人
	UserStateAgree               //已同意
	UserStateDisagree            //已反对
	UserStateOffline             //离线
	UserStateWaiting             //等待操作
)

// room user role
const (
	UserRoleMaster = 1 //庄家
	UserRoleSlave  = 2 //闲家
	UserRoleAgent  = 3 //代开房主
)

// room user ready status
const (
	UserReady   = 1
	UserUnready = 2
)

// web socket conncthion status
const (
	SocketAline = 1
	SocketClose = 2
)

const (
	CreateRoom = 1
	EnterRoom  = 2
)

const (
	ThirteenGameType = 1001
	NiuniuGameType   = 1002
	DoudizhuGameType = 1003

	ThirteenGameCost = 1.0
	NiuniuGameCost   = 1.0
	DoudizhuGameCost = 1.0

	ThirteenMaxNumber = 4
	NiuniuMaxNumber   = 5
	DoudizhuMaxNumber = 4
)

// game recovery room status
const (
	RecoveryFristInitNoReady = 1 //房间初始化玩家未准备 没有上一局
	RecoveryInitNoReady      = 2 //房间初始化玩家未准备 有上一局
	RecoveryInitReady        = 3 //房间初始化玩家已准备
	RecoveryGameStart        = 4 //游戏开始
)

const (
	MaxAgentRoomRecordCount = 20 //最大页数
	AgentRoomAllPage        = -1
	AgentRoomAllGameType    = 1
)

const (
	CostTypeGold    = 1
	CostTypeDiamond = 2
)

const (
	RoomFlag   = 1
	RoomNoFlag = 2
)
