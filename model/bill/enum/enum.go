package enum

const (
	ServiceCode = 20
)
const (
	JournalTypeInitBalance  = iota + 1 // 1:初始值
	JournalTypeRecharge                // 2:充值
	JournalTypeInvite                  //3:邀请
	JournalTypeInvited                 //4:被邀请
	JournalTypeShare                   //5:分享
	JournalTypeClubRecharge            //6：俱乐部充值
	JournalTypeMailTitem               //7：邮件奖励
)

const (
	JournalTypeThirteen       = 100101
	JournalTypeThirteenFreeze = 100102
	JournalTypeNiuniu         = 200101
	JournalTypeNiuniuFreeze   = 200102
	JournalTypeDoudizhu       = 300101
	JournalTypeDoudizhuFreeze = 300102
	JournalTypeFourCard       = 400101
	JournalTypeFourCardFreeze = 400102
)

const (
	UserBalanceTableName = "balances"
	JournalTableName     = "journals"
)

const (
	TypeGold    = 1 //金币
	TypeDiamond = 2 //钻石
)

//const (
//	SystemOpUserID = 100000
//	SystemOpUserName = "财神棋牌客服"
//)

const (
	OrderFail    = 0
	OrderSuccess = 1
	OrderExist   = 2
)

const (
	DefaultChannel = "system"
)

const (
	TypeUser = 1
	TypeClub = 2
)

const (
	MailRecharge = 1001
)

const (
	NameGold    = "金币"
	NameDiamond = "钻石"
)
