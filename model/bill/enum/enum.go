package enum

const (
	JournalTypeInitBalance = iota + 1
	JournalTypeDeposito
	JournalTypeDash
	JournalTypeRoom
)

const (
	UserBalanceTableName = "balances"
	JournalTableName     = "journals"
)
