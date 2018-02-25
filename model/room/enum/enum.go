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
	MaxRecordCount         = 10
	RoomDelayMinutes       = 5.0
	RoomCodeMax            = 999999
	RoomCodeMin            = 100000
	RoomGiveupCleanMinutes = 5.0
	ShuffleDelaySeconds    = 3
)

//db config
const (
	RoomTableName       = "rooms"
	PlayerRoomTableName = "player_rooms"
)

// room status desc
const (
	RoomStatusInit            = 110 //建立房间
	RoomStatusAllReady        = 120 //房间满员并且全部确认准备
	RoomStatusStarted         = 130 //游戏开始
	RoomStatusShuffle         = 140 //洗牌
	RoomStatusReInit          = 150 //一局游戏结束，等待下一轮确认开始
	RoomStatusDelay           = 160 //房间到达游戏最大局数，留存一段时间等待同房续费
	RoomStatusDone            = 170 //房间正常结束
	RoomStatusDestroy         = 180 //房间非正常结束(所有人员离开)
	RoomStatusGiveUp          = 190 //房间投票放弃
	RoomStatusOverTimeClean   = 200 //房间长时间无人操作被超时清除
	RoomStatusDiamondNoEnough = 210 //房间开始时付费人水晶不足 取消房间数据
)

// room type desc
const (
	RoomTypeNom   = 1 //普通房
	RoomTypeAgent = 2 //代理开房
	RoomTypeClub  = 3 //俱乐部
	RoomTypeGold  = 4 //金币场
)

// give up vote status
const (
	GiveupStatusAgree    = 1 //同意投降
	GiveupStatusDisAgree = 2 //不同意投降
	GiveupStatusWairting = 3 //继续等待投票
)

const (
	NoGiveUp   = 1 //游戏正常进行
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
	UserReady   int32 = 1
	UserUnready int32 = 2
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
	FourCardGameType = 1004
	TowCardGameType  = 1005

	ThirteenGameCost = 1.0
	NiuniuGameCost   = 1.0
	DoudizhuGameCost = 1.0
	FourcardGameCost = 1.0
	TowcardGameCost  = 1.0

	ThirteenMaxNumber = 8
	NiuniuMaxNumber   = 10
	DoudizhuMaxNumber = 4
	FourcardMaxNumber = 8
	TowcardMaxNumber  = 10
)

// game recovery room status
const (
	RecoveryFristInitNoReady = 110 //房间初始化玩家未准备 没有上一局
	RecoveryInitNoReady      = 120 //房间初始化玩家未准备 有上一局
	RecoveryInitReady        = 130 //房间初始化玩家已准备
	RecoveryGameStart        = 140 //游戏开始
	RecoveryGameDone         = 150 //游戏开始
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

const (
	ConsumeOpen      = 110001 //消费开关
	AgentConsumeOpen = 110002 //代开房消费开关
	ClubConsumeOpen  = 110003 //俱乐部消费开关
)

const (
	Player = 1
	Robot  = 2
)

var GameType = []int32{1001, 1002, 1003, 1004}
