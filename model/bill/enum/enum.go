package enum

const (
	JournalTypeInitBalance = iota + 1
	JournalTypeDeposito
	JournalTypeDash
	JournalTypeGive
	JournalTypeRecharge
	JournalTypeInvite
	JournalTypeInvited
	JournalTypeShare
	JournalTypeRoom
	JournalTypeThirteen
	JournalTypeNiuniu
)

const (
	UserBalanceTableName = "balances"
	JournalTableName     = "journals"
)

const (
	TypeGold = 1
	TypeDiamond =2
)

const (
	SystemOpUserID = 1000000001
)

const (
	OrderFail    = 0
	OrderSuccess = 1
	OrderExist   = 2
)

const (
	DefaultChannel = "system"
)
