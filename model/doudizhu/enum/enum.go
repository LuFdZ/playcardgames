package enum

const (
	LoopTime = 500
)
const (
	GetBankerCountDown  = 20
	SubmitCardCountDown = 20
)

const (
	DDZBankerTypeCall = 1 //叫地主
	DDZBankerTypeGet  = 2 //抢地主
)

const (
	DDZCallBanker = 1 //叫地主
	DDZGetBanker  = 2 //抢地主
	DDZNoBanker   = 3 //不叫
)

const (
	Dizhu   = 1
	NongMin = 2
)

const (
	DDZBankerStatusReStart  = 1 //无人叫庄 重新发牌
	DDZBankerStatusContinue = 2 //继续叫庄
	DDZBankerStatusFinish   = 3 //叫庄结束选出庄家
)

const (
	GameStatusInit       = 1 //游戏初始化发牌开始抢庄
	GameStatusSubmitCard = 2 //游戏开始交牌
	GameStatusStarted    = 3 //计算结果
	GameStatusDone       = 4 //游戏已结算
	GameStatusGiveUp     = 5 //游戏被放弃
)

const (
	UserStatusInit       = 1 //获得发牌
	UserStatusGetBanker  = 2 //已抢庄
	UserStatusSubmitCard = 4 //已提交
	UserStatusDone       = 5 //结束游戏
)

const (
	Success = 1
	fail    = 2
)
