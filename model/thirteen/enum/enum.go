package enum

const (
	LoopTime       = 500
	MaxRecordCount = 50
	LogTime        = 20
	ServiceCode    = 50
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
	TimesDefault = 1 //当庄模式 固定上庄
)

var GroupTypeName = [...]string{"Single", "Couple", "TwoCouple", "Three", "Straight", "Flush", "ThreeCouple", "Four", "FlushStraight", "Shoot", "AllShoot"}

//var TestCardListMap = [][][]string{
//	[][]string{[]string{"1_3","1_4","1_5"},[]string{"1_6","1_7","1_8","1_9","1_10"},[]string{"1_11","1_12","1_13","2_1","2_2"}},
//	[][]string{[]string{"1_3","1_4","1_5"},[]string{"1_6","1_7","1_8","1_9","1_10"},[]string{"1_11","1_12","1_13","2_1","2_2"}},
//	[][]string{[]string{"1_3","1_4","1_5"},[]string{"1_6","1_7","1_8","1_9","1_10"},[]string{"1_11","1_12","1_13","2_1","2_2"}},
//	[][]string{[]string{"1_3","1_4","1_5"},[]string{"1_6","1_7","1_8","1_9","1_10"},[]string{"1_11","1_12","1_13","2_1","2_2"}},
//	[][]string{[]string{"1_3","1_4","1_5"},[]string{"1_6","1_7","1_8","1_9","1_10"},[]string{"1_11","1_12","1_13","2_1","2_2"}},
//}
var TestCardListMap = [][][]string{
	[][]string{[]string{"3_11","2_12","2_13"},[]string{"4_3","4_3","4_6","3_6","4_10"},[]string{"3_1","2_1","4_1","4_1","2_7"}},
	[][]string{[]string{"2_5","2_5","1_1"},[]string{"4_10","2_10","3_4","2_2","1_2"},[]string{"4_9","3_9","2_9","2_9","1_9"}},
	[][]string{[]string{"3_10","4_11","4_5"},[]string{"4_13","1_13","3_12","3_12","2_12"},[]string{"1_3","2_3","2_3","3_3","4_2"}},
	[][]string{[]string{"4_12","4_13","1_1"},[]string{"3_3","1_5","4_6","4_4","3_2"},[]string{"1_8","1_9","1_7","1_10","1_11"}},
	[][]string{[]string{"2_10","2_11","2_4"},[]string{"4_2","3_5","4_5","2_8","3_8"},[]string{"1_2","1_4","1_7","1_11","1_12"}},
	[][]string{[]string{"1_10","4_12","3_1"},[]string{"2_2","3_5","4_9","3_13","3_13"},[]string{"3_11","2_11","4_11","1_4","4_4"}},
	[][]string{[]string{"3_9","3_10","1_12"},[]string{"1_3","1_5","2_6","2_7","4_7"},[]string{"1_8","4_8","2_8","3_4","2_4"}},
	[][]string{[]string{"1_13","2_13","2_1"},[]string{"3_7","3_7","4_7","3_8","4_8"},[]string{"1_6","1_6","2_6","3_6","3_2"}},
}