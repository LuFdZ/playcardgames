package enum

const (
	JournalTypeInitBalance  = iota + 1
	JournalTypeRecharge
	JournalTypeInvite
	JournalTypeInvited
	JournalTypeShare
	JournalTypeClubRecharge
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
	TypeGold    = 1
	TypeDiamond = 2
)

const (
	SystemOpUserID = 100000
	SystemOpUserName = "财神棋牌客服"
)

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
