package enum

const (
	LoopTime             = 500
	ClubMemberLimit      = 100
	ClubTableName        = "clubs"
	ClubJournalTableName = "club_journals"
	MaxRecordCount       = 20
	ClubCoinInit         = 10000000
)

const (
	JournalTypeClubAddMemberClubCoin     int32 = 1
	JournalTypeClubMemberClubCoinOfferUp int32 = 2
	JournalTypeClubGame                  int32 = 3
	JournalTypeClubGameCostBack          int32 = 4
	JournalTypeClubRecharge              int32 = 6
)

const (
	ClubMember = 1
	ClubMaster = 2
)

const (
	ClubStatusExamine = -1
	ClubStatusNormal  = 1
	ClubStatusBan     = 2
	ClubStatusDel     = 3
)

const (
	UserOnline = 1
	UserUnline = 2
)

const (
	ClubMemberStatusNon         = 1
	ClubMemberStatusBan         = 2
	ClubMemberStatusLeave       = 3
	ClubMemberStatusRemoved     = 4
	ClubMemberStatusBlackList   = 5
	ClubMemberStatusDissolution = 6
)

const (
	ClubOpCreateRoom   = "ClubOpCreateRoom"
	ClubOpUpdateRoom   = "ClubOpUpdateRoom"
	ClubOpRemoveMember = "ClubOpRemoveMember"
	ClubOpCreateMember = "ClubOpCreateMember"
	ClubOpJoinClub     = "ClubOpJoinClub"
	ClubOpLeaveClub    = "ClubOpLeaveClub"
)

const (
	TypeGold     = 1
	TypeDiamond  = 2
	TypeClubCoin = 3
)

const (
	MailClubJoin     = 1101
	MailClubUnJoin   = 1102
	MailClubUnRefuse = 1103
	MailClubDelete   = 1104
)

const (
	JournalStatusInit  = 110
	MailClubStatusSure = 120
)

const (
	ProxyCreateClubLimit  = 3
	PlayerCreateClubLimit = 1
)

const (
	ProxyCreateVipRoomSettingLimit  = 9
	PlayerCreateVipRoomSettingLimit = 3
)

const (
	VipRoomSettingNew  = 100
	VipRoomSettingNon  = 1
	VipRoomSettingStop = 2
	VipRoomSettingDel  = 3
)

var JouranlTypeNameMap = map[int32]string{1: "发放奖杯", 2: "贡献奖杯", 3: "游戏结算", 4: "报名费"}
