package enum

const (
	ErrUID = -1
)

const (
	LoopTime       = 3
	MaxRecordCount = 50
)

const (
	RoomTableName = "rooms"
)

const (
	RoomStatusInit     = 1
	RoomStatusAllReady = 2
	RoomStatusStarted  = 3
	RoomStatusDone     = 4
	RoomStatusDestroy  = 5
)

const (
	UserRoleMaster = 1
	UserRoleSlave  = 2
)

const (
	UserUnready = 0
	UserReady   = 1
)
