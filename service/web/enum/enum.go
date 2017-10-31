package enum

const (
	HeartbeatTimeout = 7
	SocketAline      = 1
	SocketClose      = 2
)

const (
	MsgRoomCreate         = "RoomCreate"
	MsgRoomReady          = "RoomReady"
	MsgRoomJoin           = "RoomJoin"
	MsgRoomUnJoin         = "RoomUnJoin"
	MsgRoomResult         = "RoomResult"
	MsgRoomGiveup         = "RoomGiveup"
	MsgRoomShock          = "RoomShock"
	MsgRoomUserConnection = "UserConnection"
	MsgRoomRenewal        = "RoomRenewal"
	MsgRoomVoiceChat      = "RoomVoiceChat"
	MsgRoomExist         = "RoomExist"
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

const (
	MsgHearbeat = "Hearbeat"
)
