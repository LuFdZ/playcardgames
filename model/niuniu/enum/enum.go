package enum

const (
	LoopTime       = 1
	MaxRecordCount = 50
	GetBankerTime  = 5.0
	SetBetTime     = 5.0
	SubmitCardTime = 5.0
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
	GameStatusAllSetBet  = 4 //游戏所有玩家都已下注结束
	GameStatusSubmitCard = 5 //游戏开始交牌
	GameStatusStarted    = 6 //计算结果
	GameStatusDone       = 7 //游戏已结算
	GameStatusGiveUp     = 8 //游戏被放弃
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
	UserReady   = 1
	UserUnready = 2
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

var (
	BankerScoreMap map[int32]int32
	BetScoreMap    map[int32]int32

	ToBankerScoreMap map[int32]int32
	ToBetScoreMap    map[int32]int32

	GameStatusMap map[int32]int32

	RoomParamMap map[string]int32
)
