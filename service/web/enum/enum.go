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
	MsgRoomExist          = "RoomExist"
	MsgRoomNotice         = "RoomNotice"
)

const (
	MsgThireteenGameResult = "ThirteenGameResult"
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
	MsgClubMemberJoin   = "ClubMemberJoin"
	MsgClubMemberLeave  = "ClubMemberLeave"
	MsgClubInfo         = "ClubInfo"
	MsgClubOnlineStatus = "ClubOnlineStatus"
	MsgClubRoomCreate   = "ClubRoomCreate"
	MsgClubRoomJoin     = "ClubRoomJoin"
	MsgClubRoomUnJoin   = "ClubRoomUnJoin"
	MsgClubRoomFinish   = "ClubRoomFinish"
)
