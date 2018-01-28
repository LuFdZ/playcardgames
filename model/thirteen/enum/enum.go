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
	GameStatusInit    = 110 //游戏初始化发牌
	GameStatusStarted = 120 //所有人都已提交牌
	GameStatusDone    = 130 //游戏已结算
	GameStatusGiveUp  = 140 //游戏被放弃
)

const (
	RecoveryInitNoReady = 110 //十三张初始化玩家未交牌
	RecoveryInitReady   = 120 //十三张初始化玩家已交牌
	RecoveryGameStart   = 130 //游戏结算
)

const (
	TimesDefault = 1//当庄模式 固定上庄
)

var GroupTypeName = [...]string{"Single", "Couple", "TwoCouple", "Three", "Straight", "Flush", "ThreeCouple", "Four", "FlushStraight", "Shoot", "AllShoot"}
