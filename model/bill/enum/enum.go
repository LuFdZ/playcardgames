package enum

const (
	JournalTypeInitBalance = iota + 1
	JournalTypeDeposito
	JournalTypeDash
	JournalTypeRoom
	JournalTypeGive
	JournalTypeRecharge
)

const (
	UserBalanceTableName = "balances"
	JournalTableName     = "journals"
)

const (
	SystemOpUserID = 1000000001
)

const (
	OrderFail    = 0
	OrderSuccess = 1
	OrderExist   = 2
)
