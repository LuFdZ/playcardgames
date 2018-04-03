package mod

import (
	mdtime "playcards/model/time"
	cacheuser "playcards/model/user/cache"
	pbclub "playcards/proto/club"
	"encoding/json"
	"time"
	//cacheclub "playcards/model/club/cache"
	//"playcards/model/user/mod"
	"github.com/jinzhu/gorm"
)

type Club struct {
	ClubID       int32         `gorm:"primary_key"`
	ClubName     string
	Status       int32
	CreatorID    int32
	CreatorProxy int32
	ClubRemark   string
	Icon         string
	ClubParam    string
	Diamond      int64
	Gold         int64
	ClubCoin     int64
	Notice       string
	SettingParam string
	MemberNumber int32         `gorm:"-"`
	Setting      *SettingParam `gorm:"-"`
	CreatedAt    *time.Time
	UpdatedAt    *time.Time
}

type ClubMember struct {
	MemberId  int32 `gorm:"primary_key"`
	ClubID    int32
	UserID    int32
	Role      int32
	Status    int32
	ClubCoin  int64
	Online    int32 `gorm:"-"`
	CreatedAt *time.Time
	UpdatedAt *time.Time
}

type ClubJournal struct {
	ID           int32 `gorm:"primary_key"`
	ClubID       int32
	AmountType   int32
	Amount       int64
	AmountBefore int64
	AmountAfter  int64
	Type         int32
	Foreign      int64
	OpUserID     int64
	Status       int32
	CreatedAt    *time.Time
	UpdatedAt    *time.Time
}

type VipRoomSetting struct {
	ID           int32 `gorm:"primary_key"`
	Name         string
	ClubID       int32
	UserID       int32
	RoomType     int32
	MaxNumber    int32
	RoundNumber  int32
	SubRoomType  int32
	GameParam    string
	Status       int32
	GameType     int32
	SettingParam string
	CreatedAt    *time.Time
	UpdatedAt    *time.Time
}

type SettingParam struct {
	CostType          int32
	CostValue         int32
	ClubCoinBaseScore int64
	CostRange         int32
}

func ClubFromProto(pclub *pbclub.Club) *Club {
	out := &Club{
		ClubID:       pclub.ClubID,
		ClubName:     pclub.ClubName,
		Status:       pclub.Status,
		CreatorID:    pclub.CreatorID,
		CreatorProxy: pclub.CreatorProxy,
		ClubRemark:   pclub.ClubRemark,
		Icon:         pclub.Icon,
		ClubParam:    pclub.ClubParam,
		Notice:       pclub.Notice,
		CreatedAt:    mdtime.TimeFromProto(pclub.CreatedAt),
		UpdatedAt:    mdtime.TimeFromProto(pclub.UpdatedAt),
	}
	if pclub.SettingParam != nil {
		out.Setting = SettingParamFromProto(pclub.SettingParam)
	}

	return out
}

func SettingParamFromProto(pbsp *pbclub.SettingParam) *SettingParam {
	return &SettingParam{
		CostType:          pbsp.CostType,
		CostValue:         pbsp.CostValue,
		ClubCoinBaseScore: pbsp.ClubCoinBaseScore,
		CostRange:         pbsp.CostRange,
	}
}

func ClubMemberFromProto(pcm *pbclub.ClubMember) *ClubMember {
	return &ClubMember{
		MemberId:  pcm.MemberID,
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
		Notice:       club.Notice,
		ClubCoin:     club.ClubCoin,
		SettingParam: club.Setting.ToProto(),
		CreatedAt:    mdtime.TimeToProto(club.CreatedAt),
		UpdatedAt:    mdtime.TimeToProto(club.UpdatedAt),
		MemberNumber: club.MemberNumber,
	}
}

func (mcm *ClubMember) ToProto() *pbclub.ClubMember {
	mCm := &pbclub.ClubMember{
		ClubID:    mcm.ClubID,
		UserID:    mcm.UserID,
		Status:    mcm.Status,
		Role:      mcm.Role,
		ClubCoin:  mcm.ClubCoin,
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

func (club *ClubJournal) ToProto() *pbclub.ClubJournal {
	return &pbclub.ClubJournal{
		ID:           club.ID,
		ClubID:       club.ClubID,
		AmountType:   club.AmountType,
		Amount:       club.Amount,
		AmountBefore: club.AmountBefore,
		AmountAfter:  club.AmountAfter,
		Type:         club.Type,
		Foreign:      club.Foreign,
		OpUserID:     club.OpUserID,
		Status:       club.Status,
		CreatedAt:    mdtime.TimeToProto(club.CreatedAt),
		UpdatedAt:    mdtime.TimeToProto(club.UpdatedAt),
	}
}

func (mdsp *SettingParam) ToProto() *pbclub.SettingParam {
	if mdsp == nil {
		return nil
	}
	return &pbclub.SettingParam{
		CostType:          mdsp.CostType,
		CostValue:         mdsp.CostValue,
		ClubCoinBaseScore: mdsp.ClubCoinBaseScore,
		CostRange:         mdsp.CostRange,
	}
}

func (vrs *VipRoomSetting) ToProto() *pbclub.VipRoomSetting {
	//clubMemberNumber, _ := cacheclub.CountClubMemberHKeys(club.ClubID)
	return &pbclub.VipRoomSetting{
		VipRoomSettingID: vrs.ID,
		Name:             vrs.Name,
		ClubID:           vrs.ClubID,
		UserID:           vrs.UserID,
		RoomType:         vrs.RoomType,
		MaxNumber:        vrs.MaxNumber,
		RoundNumber:      vrs.RoundNumber,
		SubRoomType:      vrs.SubRoomType,
		GameParam:        vrs.GameParam,
		Status:           vrs.Status,
		GameType:         vrs.GameType,
		SettingParam:     vrs.SettingParam,
	}
}

func VipRoomSettingFromProto(pbvrs *pbclub.VipRoomSetting) *VipRoomSetting {
	out := &VipRoomSetting{
		ID:           pbvrs.VipRoomSettingID,
		Name:         pbvrs.Name,
		ClubID:       pbvrs.ClubID,
		UserID:       pbvrs.UserID,
		RoomType:     pbvrs.RoomType,
		MaxNumber:    pbvrs.MaxNumber,
		RoundNumber:  pbvrs.RoundNumber,
		SubRoomType:  pbvrs.SubRoomType,
		GameParam:    pbvrs.GameParam,
		Status:       pbvrs.Status,
		GameType:     pbvrs.GameType,
		SettingParam: pbvrs.SettingParam,
	}
	return out
}

func (c *Club) BeforeUpdate(scope *gorm.Scope) error {
	if c.Setting == nil {
		return nil
	}
	c.MarshalSettingParam()
	scope.SetColumn("setting_param", c.SettingParam)
	return nil
}

func (c *Club) BeforeCreate(scope *gorm.Scope) error {
	if c.Setting == nil {
		return nil
	}
	c.MarshalSettingParam()
	scope.SetColumn("setting_param", c.SettingParam)
	return nil
}

func (c *Club) AfterFind() error {
	err := c.UnmarshalSettingParam()
	if err != nil {
		return err
	}
	return nil
}

func (c *Club) MarshalSettingParam() error {
	if c.Setting == nil {
		return nil
	}
	data, _ := json.Marshal(&c.Setting)
	c.SettingParam = string(data)
	return nil
}

func (c *Club) UnmarshalSettingParam() error {
	var out *SettingParam
	if len(c.SettingParam) > 0 {
		if err := json.Unmarshal([]byte(c.SettingParam), &out); err != nil {
			return err
		}
		c.Setting = out
	}
	return nil
}
