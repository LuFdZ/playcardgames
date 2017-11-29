package enum

const (
	LoopTime        = 500
	ClubMemberLimit = 100
	ClubJournalTableName = "club_journals"
)

const (
	ClubMember = 1
	ClubMaster = 2
)
const (
	ClubStatusExamine = -1
	ClubStatusNormal  = 1
	ClubStatusBan     = 2
)

const (
	UserOnline = 1
	UserUnline = 2
)

const (
	ClubMemberStatusNon       = 1
	ClubMemberStatusBan       = 2
	ClubMemberStatusLeave     = 3
	ClubMemberStatusRemoved   = 4
	ClubMemberStatusBlackList = 5
)

const (
	ClubOpCreateRoom   = "ClubOpCreateRoom"
	ClubOpUpdateRoom   = "ClubOpUpdateRoom"
	ClubOpRemoveMember = "ClubOpRemoveMember"
	ClubOpCreateMember = "ClubOpCreateMember"
	ClubOpJoinClub     = "ClubOpJoinClub"
	ClubOpLeaveClub     = "ClubOpLeaveClub"
)


const (
	TypeGold = 1
	TypeDiamond =2
)