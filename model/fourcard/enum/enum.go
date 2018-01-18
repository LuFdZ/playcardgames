package enum

const (
	LoopTime       = 500
	MaxRecordCount = 50
	LogTime        = 20
	SetBetTime     = 17.0
	SubmitCardTime = 1200.0
)

const (
	FourCardTableName = "fourcards"
)

const (
	Success = 1
	Fail    = 2
)

const (
	GameID = 1004
)

const (
	GameStatusInit          = 1 //游戏初始化发牌开始下注
	GameStatusAllBet        = 2 //所有人下注完成开始定序
	GameStatusOrdered       = 3 //顺序确定发牌
	GameStatusSubmitCard    = 4 //开始交牌
	GameStatusAllSubmitCard = 5 //所有人都已提交牌
	GameStatusDone          = 6 //游戏已结算
	GameStatusGiveUp        = 7 //游戏被放弃
)

const (
	UserStatusInit       = 1 //游戏初始化
	UserStatusSetBet     = 2 //已下注
	UserStatusSubmitCard = 3 //已提交
)

const (
	GetBetStatus        = 1
	GetSubmitCardStatus = 2
)

const (
	RecoveryInitNoReady = 1 //十三张初始化玩家未交牌
	RecoveryInitReady   = 2 //十三张初始化玩家已交牌
	RecoveryGameStart   = 3 //游戏结算
)

const (
	TimesDefault = 3 //当庄模式 固定上庄

)

var FourCardCardType = [...]int32{10, 20, 30, 40, 50, 60, 70, 80, 90,
	100, 110}

var BetScoreMap = map[int32]int32{1: 5, 2: 10, 3: 15, 4: 20, 5: 25}
