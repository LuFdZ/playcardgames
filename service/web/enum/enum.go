package enum

const (
	HeartbeatTimeout = 7
	SocketAline      = 1
	SocketClose      = 2
)

const (
	MsgRoomStatusChange   = "RoomStatusChange"
	MsgRoomReady          = "RoomReady"
	MsgRoomUnReady        = "RoomUnReady"
	MsgRoomJoin           = "RoomJoin"
	MsgRoomUnJoin         = "RoomUnJoin"
	MsgRoomResult         = "RoomResult"
	MsgRoomGiveup         = "RoomGiveup"
	MsgRoomShock          = "RoomShock"
	MsgRoomUserConnection = "UserConnection"
	MsgRoomRenewal        = "RoomRenewal"
)

const (
	MsgThireteenGameResult = "ThirteenGameResult"
	MsgThireteenSurrender  = "ThirteenSurrender"
	MsgThireteenGameReady  = "ThirteenGameReady"
	MsgThireteenGameStart  = "ThirteenGameStart"
)

const (
	MsgNiuniuGameResult = "NiuniuGameResult"
	MsgNiuniuBeBanker   = "NiuniuBeBanker"
	MsgNiuniuSetBet     = "NiuniuSetBet"
	MsgNiuniuAllBet     = "NiuniuAllBet"
	MsgNiuniuGameReady  = "NiuniuGameReady"
	MsgNiuniuGameStart  = "NiuniuGameStart"
	MsgNiuniuCountDown  = "NiuniuCountDown"
)

const (
	MsgBillChange = "BillChange"
)
