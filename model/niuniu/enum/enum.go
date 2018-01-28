package enum

const (
	LoopTime       = 500
	GetBankerTime  = 12.0
	SetBetTime     = 17.0
	SubmitCardTime = 12.0
	MinSetBet      = 5
)

const (
	NiuniuTableName = "niunius"
)

const (
	GameID = 1002
)

const (
	GameStatusInit       = 110 //游戏初始化发牌开始抢庄
	GameStatusGetBanker  = 120 //游戏抢庄结束
	GameStatusSetBet     = 130 //游戏开始下注抢
	GameStatusAllSetBet  = 140 //游戏下注结束
	GameStatusSubmitCard = 150 //游戏开始交牌
	GameStatusStarted    = 160 //计算结果
	GameStatusDone       = 170 //游戏已结算
	GameStatusGiveUp     = 180 //游戏被放弃
	GameStatusCountDown  = 190 //游戏倒计时
)

const (
	UserStatusInit       = 110 //获得发牌
	UserStatusGetBanker  = 120 //已抢庄
	UserStatusSetBet     = 130 //已下注
	UserStatusSubmitCard = 140 //已提交
	UserStatusDone       = 150 //结束游戏
)

const (
	Banker = 1
	Player = 2
)

const (
	Success = 1
	Fail    = 2
)

const (
	GetBankerStatus     = 1
	GetBetStatus        = 2
	GetSubmitCardStatus = 3
)

const (
	BankerNoNiu   = 1 //无牛下庄
	BankerTurns   = 2 //轮流上庄
	BankerSeeCard = 3 //看牌上庄
	BankerDefault = 4 //固定庄家
)

// var (
// 	BankerScoreMap map[int32]int32
// 	BetScoreMap    map[int32]int32

// 	ToBankerScoreMap map[int32]int32
// 	ToBetScoreMap    map[int32]int32

// 	GameStatusMap map[int32]int32

// 	RoomParamMap map[string]int32
// )

var NiuniuCardType = [...]string{"0", "1", "2", "3", "4", "5", "6", "7", "8",
	"9", "10", "11", "12", "13"}

var BankerScoreMap = map[int32]int32{1: 0, 2: 1, 3: 2, 4: 3, 5: 4}

var BetScoreMap = map[int32]int32{1: 5, 2: 10, 3: 15, 4: 20, 5: 25}

var ToBankerScoreMap = map[int32]int32{0: 1, 1: 2, 2: 3, 3: 4, 4: 5}

var ToBetScoreMap = map[int32]int32{110: 110, 120: 120, 130: 120, 140: 130, 150: 130, 160: 140, 170: 140, 180: 140}
