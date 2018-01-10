package enum

const (
	JournalTypeInitBalance      = 1 //iota + 1
	JournalTypeRecharge         = 2
	JournalTypeInvite           = 3
	JournalTypeInvited          = 4
	JournalTypeShare            = 5
	JournalTypeThirteen         = 6
	JournalTypeNiuniu           = 7
	JournalTypeThirteenFreeze   = 8
	JournalTypeThirteenUnFreeze = 9
	JournalTypeNiuniuFreeze     = 10
	JournalTypeNiuniuUnFreeze   = 11
	JournalTypeClubRecharge     = 12
	JournalTypeDoudizhuFreeze   = 10
	JournalTypeDoudizhuUnFreeze = 11
	JournalTypeFourCardFreeze     = 10
	JournalTypeFourCardUnFreeze   = 11
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
