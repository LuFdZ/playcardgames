package enum

const (
	LoopTime       = 3
	MaxRecordCount = 50
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

var GroupTypeName = [...]string{"Single", "Couple", "TwoCouple", "Three", "Straight", "Flush", "ThreeCouple", "Four", "FlushStraight", "Shoot", "AllShoot"}
