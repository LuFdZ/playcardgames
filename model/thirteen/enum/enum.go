package enum

const (
	LoopTime       = 500
	MaxRecordCount = 50
	LogTime        = 20
)

const (
	ThirteenTableName = "thirteens"
)

const (
	GameID = 1001
)

const (
	GameStatusInit    = 1 //游戏初始化发牌
	GameStatusStarted = 2 //所有人都已提交牌
	GameStatusDone    = 3 //游戏已结算
	GameStatusGiveUp  = 4 //游戏被放弃
)

const (
	RecoveryInitNoReady = 1 //十三张初始化玩家未交牌
	RecoveryInitReady   = 2 //十三张初始化玩家已交牌
	RecoveryGameStart   = 3 //游戏结算
)

const (
	TimesDefault = 3//当庄模式 固定上庄

)

var GroupTypeName = [...]string{"Single", "Couple", "TwoCouple", "Three", "Straight", "Flush", "ThreeCouple", "Four", "FlushStraight", "Shoot", "AllShoot"}
