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
	GameStatusInit       = 1 //游戏初始化发牌开始抢庄
	GameStatusGetBanker  = 2 //游戏抢庄结束
	GameStatusSetBet     = 3 //游戏开始下注抢
	GameStatusAllSetBet  = 4 //游戏下注结束
	GameStatusSubmitCard = 5 //游戏开始交牌
	GameStatusStarted    = 6 //计算结果
	GameStatusDone       = 7 //游戏已结算
	GameStatusGiveUp     = 8 //游戏被放弃
	GameStatusCountDown  = 9 //游戏倒计时
)

const (
	UserStatusInit       = 1 //获得发牌
	UserStatusGetBanker  = 2 //已抢庄
	UserStatusSetBet     = 3 //已下注
	UserStatusSubmitCard = 4 //已提交
	UserStatusDone       = 5 //结束游戏
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

var ToBetScoreMap = map[int32]int32{1: 1, 2: 2, 3: 2, 4: 3, 5: 3, 6: 4, 7: 4, 8: 4}
