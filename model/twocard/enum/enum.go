package enum

const (
	LoopTime       = 500
	MaxRecordCount = 50
	LogTime        = 20
	SetBetTime     = 32.0
	SubmitCardTime = 32.0
	ServiceCode    = 32
)

const (
	TowCardTableName = "twocards"
)

const (
	Success = 1
	Fail    = 2
)

const (
	GameID = 1005
)

const (
	GameStatusInit          = 110 //游戏初始化开始下注
	GameStatusAllBet        = 120 //所有人下注完成开始定序
	GameStatusOrdered       = 130 //顺序确定发牌
	GameStatusSubmitCard    = 140 //开始交牌
	GameStatusAllSubmitCard = 150 //所有人都已提交牌
	GameStatusDone          = 160 //游戏已结算
	GameStatusGiveUp        = 170 //游戏被放弃
)

const (
	UserStatusInit       = 110 //游戏初始化
	UserStatusSetBet     = 120 //已下注
	UserStatusSubmitCard = 130 //已提交
)

const (
	GetBetStatus        = 1
	GetSubmitCardStatus = 2
)

const (
	BetTypeHave = 1 //比牌模式 有下注
	BetTypeNo   = 2 //比牌模式 无下注
)

const (
	TimesDefault = 3 //当庄模式 固定上庄

)

var TwoCardCardType = [...]int32{10, 20, 30, 40, 50, 60, 70, 80, 90,
	100, 110}

//var BetScoreMap = map[int32]int32{1: 1, 2: 10, 3: 15, 4: 20, 5: 25}

var TestCardListMap = [][]string{
	{"5_21","3_3"},
	{"2_4","3_4"},
	{"3_7","4_7"},
	{"4_8","3_12"},
	{"3_2","2_8"},
	{"1_4","3_5"},
	{"2_7","4_4"},
	{"3_9","1_6"},
	{"2_10","4_10"},
	{"1_9","4_6"},
}

