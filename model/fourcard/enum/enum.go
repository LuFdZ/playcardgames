package enum

const (
	LoopTime       = 500
	MaxRecordCount = 50
	LogTime        = 20
	SetBetTime     = 32.0
	SubmitCardTime = 32.0
	ServiceCode = 16
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

var FourCardCardType = [...]int32{10, 20, 30, 40, 50, 60, 70, 80, 90,
	100, 110}

//var BetScoreMap = map[int32]int32{1: 1, 2: 10, 3: 15, 4: 20, 5: 25}

var TestCardListMap = [][][]string{
	[][]string{[]string{"3_2","1_10"},[]string{"5_21","3_3"}},
	[][]string{[]string{"1_5","4_4"},[]string{"4_8","1_8"}},
	[][]string{[]string{"3_7","3_10"},[]string{"3_9","1_12"}},
	[][]string{[]string{"3_4","4_7"},[]string{"1_2","3_8"}},
	[][]string{[]string{"1_4","3_6"},[]string{"2_10","2_8"}},
	[][]string{[]string{"1_6","2_7"},[]string{"1_9","4_10"}},
	[][]string{[]string{"3_11","2_4"},[]string{"2_6","3_12"}},
	[][]string{[]string{"4_6","3_5"},[]string{"1_11","1_7"}},
}