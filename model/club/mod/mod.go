package mod

import (
	mdtime "playcards/model/time"
	cacheuser "playcards/model/user/cache"
	pbclub "playcards/proto/club"
	"time"
	//cacheclub "playcards/model/club/cache"
	//"playcards/model/user/mod"
)

type Club struct {
	ClubID       int32 `gorm:"primary_key"`
	ClubName     string
	Status       int32
	CreatorID    int32
	CreatorProxy int32
	ClubRemark   string
	Icon         string
	ClubParam    string
	Diamond      int64
	Gold         int64
	MemberNumber int32 `gorm:"-"`
	CreatedAt    *time.Time
	UpdatedAt    *time.Time
}

type ClubMember struct {
	MemberId  int32 `gorm:"primary_key"`
	ClubID    int32
	UserID    int32
	Role      int32
	Status    int32
	Online    int32 `gorm:"-"`
	CreatedAt *time.Time
	UpdatedAt *time.Time
}

type ClubJournal struct {
	ID           int64 `gorm:"primary_key"`
	ClubID       int32
	AmountType   int32
	Amount       int64
	AmountBefore int64
	AmountAfter  int64
	Type         int32
	Foreign      int64
	OpUserID     int64
	CreatedAt    *time.Time
	UpdatedAt    *time.Time
}

func ClubFromProto(pclub *pbclub.Club) *Club {
	return &Club{
		ClubID:       pclub.ClubID,
		ClubName:     pclub.ClubName,
		Status:       pclub.Status,
		CreatorID:    pclub.CreatorID,
		CreatorProxy: pclub.CreatorProxy,
		ClubRemark:   pclub.ClubRemark,
		Icon:         pclub.Icon,
		ClubParam:    pclub.ClubParam,
		CreatedAt:    mdtime.TimeFromProto(pclub.CreatedAt),
		UpdatedAt:    mdtime.TimeFromProto(pclub.UpdatedAt),
	}
}

func ClubMemberFromProto(pcm *pbclub.ClubMember) *ClubMember {
	return &ClubMember{
		MemberId:  pcm.MemberId,
		ClubID:    pcm.ClubID,
		UserID:    pcm.UserID,
		Role:      pcm.Role,
		Status:    pcm.Status,
		CreatedAt: mdtime.TimeFromProto(pcm.CreatedAt),
		UpdatedAt: mdtime.TimeFromProto(pcm.UpdatedAt),
	}
}

func (club *Club) ToProto() *pbclub.Club {
	//clubMemberNumber, _ := cacheclub.CountClubMemberHKeys(club.ClubID)
	return &pbclub.Club{
		ClubID:       club.ClubID,
		ClubName:     club.ClubName,
		Status:       club.Status,
		CreatorID:    club.CreatorID,
		CreatorProxy: club.CreatorProxy,
		ClubRemark:   club.ClubRemark,
		Icon:         club.Icon,
		ClubParam:    club.ClubParam,
		Diamond:      club.Diamond,
		CreatedAt:    mdtime.TimeToProto(club.CreatedAt),
		UpdatedAt:    mdtime.TimeToProto(club.UpdatedAt),
		MemberNumber: club.MemberNumber,
	}
}

func (mcm *ClubMember) ToProto() *pbclub.ClubMember {
	mCm := &pbclub.ClubMember{
		ClubID: mcm.ClubID,
		UserID: mcm.UserID,
		Status: mcm.Status,
		Role:   mcm.Role,

		Online:    mcm.Online,
		CreatedAt: mdtime.TimeToProto(mcm.CreatedAt),
		UpdatedAt: mdtime.TimeToProto(mcm.UpdatedAt),
	}
	_, u := cacheuser.GetUserByID(mcm.UserID)
	if u != nil {
		mCm.Icon = u.Icon
		mCm.Nickname = u.Nickname
	}
	return mCm
}
