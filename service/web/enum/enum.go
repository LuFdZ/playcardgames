package enum

const (
	HeartbeatTimeout = 30
	SocketAline      = 1
	SocketClose      = 2
)

//服务统一编号
var (
	UserServiceCode     = 10 //用户
	BillServiceCode     = 20 //货币
	APIServiceCode      = 30
	RoomServiceCode     = 40 //房间
	ThirteenServiceCode = 50 //十三张
	NiuniuServiceCode   = 60 //牛牛
	ConfigServiceCode   = 70 //配置
	NoticeServiceCode   = 80 //通知
	ActivityServiceCode = 90 //任务活动
	CommonServiceCode   = 12 //通用
	AuthServiceCode     = 13 //权限
	ClubServiceCode     = 14 //俱乐部
	DoudizhuServiceCode = 15 //斗地主
	FourCardServiceCode = 16 //四张
	WebServiceCode      = 17 //web消息
	LogServiceCode      = 18 //日志
	SysCode             = 19 //系统
	MailServiceCode     = 21 //邮件
	GoldRoomServiceCode = 22 //金币场
	TowCardServiceCode  = 23 //两张
	RunCardServiceCode  = 24 //跑得快
)

//基本消息标签和编号
const (
	MsgSubscribeSuccess     = "SubscribeSuccess"
	MsgSubscribeSuccessCode = 191001

	MsgHeartbeat     = "ClientHeartbeat"
	MsgSubscribeCode = 191002
)

//房间消息标签和编号
const (
	MsgRoomCreate     = "RoomCreate"
	MsgRoomCreateCode = 400101

	MsgRoomReady     = "RoomReady"
	MsgRoomReadyCode = 400102

	MsgRoomJoin     = "RoomJoin"
	MsgRoomJoinCode = 400103

	MsgRoomUnJoin     = "RoomUnJoin"
	MsgRoomUnJoinCode = 400104

	MsgRoomResult     = "RoomResult"
	MsgRoomResultCode = 400105

	MsgRoomGiveup     = "RoomGiveup"
	MsgRoomGiveupCode = 400106

	MsgRoomShock     = "RoomShock"
	MsgRoomShockCode = 400107

	MsgRoomUserConnection     = "UserConnection"
	MsgRoomUserConnectionCode = 400108

	MsgRoomRenewal     = "RoomRenewal"
	MsgRoomRenewalCode = 400109

	MsgRoomVoiceChat     = "RoomVoiceChat"
	MsgRoomVoiceChatCode = 400110

	MsgRoomNotice     = "RoomNotice"
	MsgRoomNoticeCode = 400111

	MsgShuffleCard     = "ShuffleCardBro"
	MsgShuffleCardCode = 400112

	MsgRoomChat     = "RoomChatBro"
	MsgRoomChatCode = 400113

	MsgBankerList     = "RoomBankerList"
	MsgBankerListCode = 400114

	MsgUserRestore     = "UserSetRestore"
	MsgUserRestoreCode = 400115

	MsgRoomSitDown     = "RoomSitDown"
	MsgRoomSitDownCode = 400116
)

//十三张标签和编号
const (
	MsgThireteenGameResult     = "ThirteenGameResult"
	MsgThireteenGameResultCode = 500101

	MsgThireteenGameReady     = "ThirteenGameReady"
	MsgThireteenGameReadyCode = 500102

	MsgThireteenGameStart     = "ThirteenGameStart"
	MsgThireteenGameStartCode = 500103

	MsgThireteenExist     = "ThirteenExist"
	MsgThireteenExistCode = 500104
)

//牛牛标签和编号
const (
	MsgNiuniuGameResult     = "NiuniuGameResult"
	MsgNiuniuGameResultCode = 600101

	MsgNiuniuBeBanker     = "NiuniuBeBanker"
	MsgNiuniuBeBankerCode = 600102

	MsgNiuniuSetBet     = "NiuniuSetBet"
	MsgNiuniuSetBetCode = 600103

	MsgNiuniuAllBet     = "NiuniuAllBet"
	MsgNiuniuAllBetCode = 600104

	MsgNiuniuGameReady     = "NiuniuGameReady"
	MsgNiuniuGameReadyCode = 600105

	MsgNiuniuGameStart     = "NiuniuGameStart"
	MsgNiuniuGameStartCode = 600106

	MsgNiuniuExist     = "NiuniuExist"
	MsgNiuniuExistCode = 600107
)

//货币标签和编号
const (
	MsgBillChange     = "BillChange"
	MsgBillChangeCode = 200101
)

//俱乐部标签和编号
const (
	MsgClubMemberJoin     = "ClubMemberJoin"
	MsgClubMemberJoinCode = 140101

	MsgClubMemberLeave     = "ClubMemberLeave"
	MsgClubMemberLeaveCode = 140102

	MsgClubInfo     = "ClubInfo"
	MsgClubInfoCode = 140103

	MsgClubOnlineStatus     = "ClubOnlineStatus"
	MsgClubOnlineStatusCode = 140104

	MsgClubRoomCreate     = "ClubRoomCreate"
	MsgClubRoomCreateCode = 140105

	MsgClubRoomJoin     = "ClubRoomJoin"
	MsgClubRoomJoinCode = 140106

	MsgClubRoomUnJoin     = "ClubRoomUnJoin"
	MsgClubRoomUnJoinCode = 140107

	MsgClubRoomFinish     = "ClubRoomFinish"
	MsgClubRoomFinishCode = 140108

	MsgClubRoomGameStart     = "ClubRoomGameStart"
	MsgClubRoomGameStartCode = 140109

	MsgClubMemberJoinBack     = "ClubMemberJoinBack"
	MsgClubMemberJoinBackCode = 140110

	MsgClubMemberLeaveBack     = "ClubMemberLeaveBack"
	MsgClubMemberLeaveBackCode = 140111

	MsgUpdateVipRoomSetting     = "UpdateVipRoomSettingBro"
	MsgUpdateVipRoomSettingCode = 140112

	MsgUpdateClub     = "UpdateClubBro"
	MsgUpdateClubCode = 140113

	MsgClubDelete     = "ClubDeleteBro"
	MsgClubDeleteCode = 140114

	MsgAddClubCoin     = "AddClubCoin"
	MsgAddClubCoinCode = 140115

	ClubBalanceUpdate     = "ClubBalanceUpdate"
	ClubBalanceUpdateCode = 140116
)

//俱乐部标签和编号
const (
	MsgDDZGameStart     = "DoudizhuGameStart"
	MsgDDZGameStartCode = 150101

	MsgDDZBeBanker     = "DoudizhuBeBanker"
	MsgDDZBeBankerCode = 150102

	MsgDDZSubmitCard     = "DoudizhuSubmitCard"
	MsgDDZSubmitCardCode = 150103

	MsgDDZGameResult     = "DoudizhuGameResult"
	MsgDDZGameResultCode = 150104

	MsgDoudizhuExist     = "DoudizhuExist"
	MsgDoudizhuExistCode = 150105
)

//四张标签和编号
const (
	MsgFourCardGameStart     = "FourCardGameStart"
	MsgFourCardGameStartCode = 160101

	MsgFourCardSetBet     = "FourCardSetBet"
	MsgFourCardSetBetCode = 160102

	MsgFourCardGameReady     = "FourCardGameReady"
	MsgFourCardGameReadyCode = 160103

	MsgFourCardGameSubmitCard     = "FourCardSubmitCard"
	MsgFourCardGameSubmitCardCode = 160104

	MsgFourCardGameResult     = "FourCardGameResult"
	MsgFourCardGameResultCode = 160105

	MsgFourCardExist     = "FourCardExist"
	MsgFourCardExistCode = 160107
)

//邮件标签和编号
const (
	MsgNewMailNotice     = "NewMailNotice"
	MsgNewMailNoticeCode = 210101

	MsgSendSysMail     = "NewSendSysMail"
	MsgSendSysMailCode = 210102
)

//四张标签和编号
const (
	MsgTwoCardGameStart     = "TwoCardGameStart"
	MsgTwoCardGameStartCode = 230101

	MsgTwoCardSetBet     = "TwoCardSetBet"
	MsgTwoCardSetBetCode = 230102

	MsgTwoCardGameReady     = "TwoCardGameReady"
	MsgTwoCardGameReadyCode = 230103

	MsgTwoCardGameSubmitCard     = "TwoCardSubmitCard"
	MsgTwoCardGameSubmitCardCode = 230104

	MsgTwoCardGameResult     = "TwoCardGameResult"
	MsgTwoCardGameResultCode = 230105

	MsgTwoCardExist     = "TwoCardExist"
	MsgTwoCardExistCode = 230107
)

//十三张标签和编号
const (
	MsgRunCardGameResult     = "RuncardGameResult"
	MsgRunCardGameResultCode = 240101

	MsgRunCardSubmitCard     = "RuncardSubmitCard"
	MsgRunCardSubmitCardCode = 240102

	MsgRunCardGameStart     = "RuncardGameStart"
	MsgRunCardGameStartCode = 240103

	MsgRunCardExist     = "RuncardExist"
	MsgRunCardExistCode = 240104
)