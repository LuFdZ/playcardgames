package enum

const (
	JournalTypeInitBalance = iota + 1
	JournalTypeDeposito
	JournalTypeRegion
	JournalTypeZodiac
	JournalTypeDash
)

const (
	UserBalanceTableName = "balances"
	JournalTableName     = "journals"
)
