package enum

const (
	LoopTime       = 500
	GetBankerTime  = 6 //6
	SetBetTime     = 6 //6
	SubmitCardTime = 12 //12
	MinSetBet      = 1
	ServiceCode    = 60
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
	BankerSeeCard = 1 //看牌上庄
	BankerNoNiu   = 2 //无牛下庄
	BankerTurns   = 3 //轮流上庄
	BankerDefault = 4 //固定庄家
	BankerAll     = 5 //通比
)

// var (
// 	BankerScoreMap map[int32]int32
// 	BetScoreMap    map[int32]int32

// 	ToBankerScoreMap map[int32]int32
// 	ToGameStatusMap    map[int32]int32

// 	GameStatusMap map[int32]int32

// 	RoomParamMap map[string]int32
// )

var NiuniuCardType = [...]string{"0", "1", "2", "3", "4", "5", "6", "7", "8",
	"9", "10", "11", "12", "13"}

var BankerScoreMap = map[int32]int32{1: 0, 2: 1, 3: 2, 4: 3, 5: 4}

var BetScoreMap = map[int32]int32{1: 5, 2: 10, 3: 15, 4: 20, 5: 25}

var ToBankerScoreMap = map[int32]int32{0: 1, 1: 2, 2: 3, 3: 4, 4: 5}

var ToGameStatusMap = map[int32]int32{110: 110, 120: 120, 130: 120, 140: 130, 150: 130, 160: 130, 170: 140, 180: 140}

var PushOnScoreMap = map[string]int32{"1": 5, "2": 10, "3": 20}

//var TestCardListMap = [][]string{
//	{"3_5", "4_5", "1_5", "1_9", "2_9"},
//	{"2_10", "2_3", "2_1", "2_6", "2_7"},
//	{"4_8", "1_10", "4_11", "1_8", "1_6"},
//	{"1_11", "3_13", "2_12", "4_7", "1_4"},
//	{"3_9", "4_6", "3_4", "4_10", "4_4"},
//	{"1_1", "1_7", "1_3", "4_3", "3_12"},
//	{"3_3", "3_7", "4_13", "2_8", "2_2"},
//	{"2_11", "2_13", "3_8", "4_2", "2_4"},
//	{"4_1", "3_1", "2_5", "3_10", "3_2"},
//	{"1_13", "4_9", "1_2", "3_11", "4_12"},
//}

var TestCardListMap = [][]string{
	{"4_11","2_12","4_13","3_11","1_12"},
	{"1_4","4_10","4_4","1_10","3_10"},
	{"1_10","2_11","3_12","1_13","3_9"},
	{"2_6","2_11","2_7","2_13","2_9"},
	{"2_13","1_13","3_13","4_13","2_1"},
	{"3_9","3_10","3_11","3_12","3_13"},
	{"1_1","2_1","1_2","3_2","3_3"},
}