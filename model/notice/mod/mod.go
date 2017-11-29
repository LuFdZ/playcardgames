package mod

import (
	mdtime "playcards/model/time"
	pbn "playcards/proto/notice"
	utilproto "playcards/utils/proto"
	"time"
)

type Notice struct {
	NoticeID      int32 `gorm:"primary_key"`
	NoticeType    int32 `gorm:"required"`
	NoticeContent string
	Channels      string
	Versions      string
	Status        int32 `gorm:"required"`
	Description   string
	Param         string
	StartAt       *time.Time
	EndAt         *time.Time
	CreatedAt     *time.Time
	UpdatedAt     *time.Time
}

type NoticeList struct {
	List []*Notice
}

func (nl *NoticeList) ToProto() *pbn.NoticeListReply {
	out := &pbn.NoticeListReply{}
	utilproto.ProtoSlice(nl.List, &out.List)
	return out
}

func (n Notice) ToProto() *pbn.Notice {
	out := &pbn.Notice{
		NoticeID:      n.NoticeID,
		NoticeType:    n.NoticeType,
		NoticeContent: n.NoticeContent,
		Channels:      n.Channels,
		Versions:      n.Versions,
		Status:        n.Status,
		Description:   n.Description,
		Param:         n.Param,
		StartAt:       mdtime.TimeToProto(n.StartAt),
		EndAt:         mdtime.TimeToProto(n.EndAt),
		CreatedAt:     mdtime.TimeToProto(n.CreatedAt),
		UpdatedAt:     mdtime.TimeToProto(n.UpdatedAt),
	}
	return out
}

func NoticeFromProto(n *pbn.Notice) *Notice {
	if n == nil {
		return &Notice{}
	}
	return &Notice{
		NoticeID:      n.NoticeID,
		NoticeType:    n.NoticeType,
		NoticeContent: n.NoticeContent,
		Channels:      n.Channels,
		Versions:      n.Versions,
		Status:        n.Status,
		Description:   n.Description,
		Param:         n.Param,
		StartAt:       mdtime.TimeFromProto(n.StartAt),
		EndAt:         mdtime.TimeFromProto(n.EndAt),
		CreatedAt:     mdtime.TimeFromProto(n.CreatedAt),
		UpdatedAt:     mdtime.TimeFromProto(n.UpdatedAt),
	}
}
