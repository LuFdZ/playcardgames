package enum

const (
	LoopTime       = 500
	MaxRecordCount = 50
	LogTime        = 20
	SubmitCardTime = 32000.0
	ServiceCode    = 32
	BoomType       = 10
)

const (
	RunCardTableName = "runcards"
)

const (
	Success = 1
	Fail    = 2
)

const (
	GameID = 1006
)

var RunCardCardType = [...]int32{10, 20, 30, 40, 50, 60, 70, 80, 90,
	100, 110}

var RunCardBoomType = [...]string{"3303", "20", "30", "40", "50"}

const (
	GameStatusInit           = 110 //游戏初始化游戏开始交牌
	GameStatusSubmitCardOver = 130 //交牌结束
	GameStatusDone           = 130 //游戏已结算
	GameStatusGiveUp         = 140 //游戏被放弃
)

const (
	UserStatusInit = 110 //游戏初始化发牌
	UserStatusDone = 120 //游戏已结算
)

const (
	CompareSuccess = "1"
	CompareFail    = "2"
)

const (
	CardTypeError     = "-1"
	CardValueError    = "2"
	HaveToSubmitError = "3"
)

const (
	IsSpring = 1
	IsNom    = 2
	IsClose  = 3
)
