package enum

const (
	LoopTime       = 500
	MaxRecordCount = 50
	LogTime        = 20
	SetBetTime     = 17.0
	SubmitCardTime = 12.0
)

const (
	FourCardTableName = "fourcards"
)

const (
	GameID = 1004
)

const (
	GameStatusInit          = 1 //游戏初始化发牌开始下注
	GameStatusAllBet        = 2 //所有人下注完成开始定序
	GameStatusOrdered       = 3 //顺序确定开始交牌
	GameStatusAllSubmitCard = 4 //所有人都已提交牌
	GameStatusDone          = 5 //游戏已结算
	GameStatusGiveUp        = 6 //游戏被放弃
)

const (
	UserStatusInit       = 1 //游戏初始化
	UserStatusSetBet     = 1 //已下注
	UserStatusSubmitCard = 4 //已提交
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

var FourCardCardType = [...]string{"0", "1", "2", "3", "4", "5", "6", "7", "8",
	"9", "10", "11", "12", "13"}

var BetScoreMap = map[int32]int32{1: 5, 2: 10, 3: 15, 4: 20, 5: 25}
