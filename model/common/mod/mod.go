package mod

import (
	mdtime "playcards/model/time"
	cacheuser "playcards/model/user/cache"
	pbcon "playcards/proto/common"
	"time"
)

type BlackList struct {
	BlackID   int32 `gorm:"primary_key"`
	Type      int32
	OriginID  int32
	TargetID  int32
	OpID      int32
	Status    int32
	CreatedAt *time.Time
	UpdatedAt *time.Time
}

type Examine struct {
	ExamineID   int32 `gorm:"primary_key"`
	Type        int32
	ApplicantID int32
	AuditorID   int32
	Status      int32
	OpID        int32
	CreatedAt   *time.Time
	UpdatedAt   *time.Time
}

func BlackListFromProto(pbbl *pbcon.BlackList) *BlackList {
	return &BlackList{
		BlackID:   pbbl.BlackID,
		Type:      pbbl.Type,
		OriginID:  pbbl.OriginID,
		TargetID:  pbbl.TargetID,
		OpID:      pbbl.OpID,
		Status:    pbbl.Status,
		CreatedAt: mdtime.TimeFromProto(pbbl.CreatedAt),
		UpdatedAt: mdtime.TimeFromProto(pbbl.UpdatedAt),
	}
}

func (mbl *BlackList) ToProto() *pbcon.BlackList {
	pbBl := &pbcon.BlackList{
		BlackID:   mbl.BlackID,
		Type:      mbl.Type,
		OriginID:  mbl.OriginID,
		TargetID:  mbl.TargetID,
		OpID:      mbl.OpID,
		Status:    mbl.Status,
		CreatedAt: mdtime.TimeToProto(mbl.CreatedAt),
		UpdatedAt: mdtime.TimeToProto(mbl.UpdatedAt),
	}

	_, u := cacheuser.GetUserByID(pbBl.TargetID)
	if u != nil {
		pbBl.Icon = u.Icon
		pbBl.Nickname = u.Nickname
	}
	return pbBl
}

func ExamineFromProto(pbme *pbcon.Examine) *Examine {
	return &Examine{
		ExamineID:   pbme.ExamineID,
		Type:        pbme.Type,
		ApplicantID: pbme.ApplicantID,
		AuditorID:   pbme.AuditorID,
		Status:      pbme.Status,
		OpID:        pbme.OpID,
		CreatedAt:   mdtime.TimeFromProto(pbme.CreatedAt),
		UpdatedAt:   mdtime.TimeFromProto(pbme.UpdatedAt),
	}
}

func (me *Examine) ToProto() *pbcon.Examine {
	pbMe := &pbcon.Examine{
		ExamineID:   me.ExamineID,
		Type:        me.Type,
		ApplicantID: me.ApplicantID,
		AuditorID:   me.AuditorID,
		Status:      me.Status,
		OpID:        me.OpID,
		CreatedAt:   mdtime.TimeToProto(me.CreatedAt),
		UpdatedAt:   mdtime.TimeToProto(me.UpdatedAt),
	}

	_, u := cacheuser.GetUserByID(me.ApplicantID)
	if u != nil {
		pbMe.Icon = u.Icon
		pbMe.Nickname = u.Nickname
	}
	return pbMe
}
