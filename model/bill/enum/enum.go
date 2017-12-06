package enum

const (
	JournalTypeInitBalance = iota + 1
	JournalTypeRecharge
	JournalTypeInvite
	JournalTypeInvited
	JournalTypeShare
	JournalTypeThirteen
	JournalTypeNiuniu
	JournalTypeThirteenFreeze
	JournalTypeThirteenUnFreeze
	JournalTypeNiuniuFreeze
	JournalTypeNiuniuUnFreeze
	JournalTypeClubRecharge
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
	SystemOpUserID = 100000
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
