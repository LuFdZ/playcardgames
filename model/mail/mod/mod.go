package mod

import (
	"encoding/json"
	pbmail "playcards/proto/mail"
	//cachegame "playcards/model/mail/cache"
	enumgame "playcards/model/mail/enum"
	mdtime "playcards/model/time"
	utilproto "playcards/utils/proto"
	"time"
	"github.com/jinzhu/gorm"
)

type MailInfo struct {
	MailID       int32 //`gorm:"primary_key"`
	MailName     string
	MailTitle    string
	MailContent  string
	MailType     int32
	Status       int32
	ItemList     string
	ItemModeList []*ItemModel
	CreatedAt    *time.Time
	UpdatedAt    *time.Time
}

type MailSendLog struct {
	LogID      int32 `gorm:"primary_key"`
	MailID     int32
	SendID     int32
	MailType   int32
	Status     int32
	MailStr    string
	SendCount  int32
	TotalCount int32
	MailInfo   *MailInfo
	CreatedAt  *time.Time
	UpdatedAt  *time.Time
}

type PlayerMail struct {
	LogID     int32 `gorm:"primary_key"`
	SendLogID int32
	MailID    int32
	UserID    int32
	SendID    int32
	MailType  int32
	Status    int32
	HaveItem  int32
	CreatedAt *time.Time
	UpdatedAt *time.Time
}

type ItemModel struct {
	MainType int32
	SubType  int32
	ItemID   int32
	Count    int64
}

func (mi *MailInfo) ToProto() *pbmail.MailInfo {
	out := &pbmail.MailInfo{
		MailID:      mi.MailID,
		MailName:    mi.MailName,
		MailTitle:   mi.MailTitle,
		MailContent: mi.MailContent,
		MailType:    mi.MailType,
		CreatedAt:   mdtime.TimeToProto(mi.CreatedAt),
		UpdatedAt:   mdtime.TimeToProto(mi.UpdatedAt),
	}
	utilproto.ProtoSlice(mi.ItemList, &out.ItemList)
	return out
}

func (msl *MailSendLog) ToProto() *pbmail.MailSendLog {
	return &pbmail.MailSendLog{
		LogID:      msl.LogID,
		MailID:     msl.MailID,
		SendID:     msl.SendID,
		MailType:   msl.MailType,
		Status:     msl.Status,
		MailInfo:   msl.MailInfo.ToProto(),
		TotalCount: msl.TotalCount,
		SendCount:  msl.SendCount,
		CreatedAt:  mdtime.TimeToProto(msl.CreatedAt),
		UpdatedAt:  mdtime.TimeToProto(msl.UpdatedAt),
	}
}

func (im *ItemModel) ToProto() *pbmail.ItemModel {
	return &pbmail.ItemModel{
		MainType: im.MainType,
		SubType:  im.SubType,
		ItemID:   im.ItemID,
		Count:    im.Count,
	}
}

func (pm *PlayerMail) ToProto() *pbmail.PlayerMail {
	out := &pbmail.PlayerMail{
		LogID:     pm.LogID,
		SendLogID: pm.SendLogID,
		UserID:    pm.UserID,
		MailType:  pm.MailType,
		Status:    pm.Status,
		CreatedAt: mdtime.TimeToProto(pm.CreatedAt),
	}
	//if pm.MailType == enumgame.MailTypeSysMode {
	//	msl, err := cachegame.GetMailSendLog(pm.SendLogID)
	//	if err != nil {
	//		return out
	//	}
	//	if msl == nil {
	//		return out
	//	}
	//	out.MailTitle = msl.MailInfo.MailTitle
	//	out.MailContent = msl.MailInfo.MailContent
	//	out.ItemList = msl.MailInfo.ItemList
	//}
	return out
}

func MailInfoFromProto(mi *pbmail.MailInfo) *MailInfo {
	out := &MailInfo{
		MailID:      mi.MailID,
		MailName:    mi.MailName,
		MailTitle:   mi.MailTitle,
		MailContent: mi.MailContent,
		MailType:    mi.MailType,
		//ItemList:    mi.ItemList,
	}
	return out
}

func ItemModelFromProto(im *pbmail.ItemModel) *ItemModel {
	out := &ItemModel{
		MainType: im.MainType,
		SubType:  im.SubType,
		ItemID:   im.ItemID,
		Count:    im.Count,
	}
	return out
}

func SendMailLogFromProto(msl *pbmail.MailSendLog) *MailSendLog {
	out := &MailSendLog{
		MailID:   msl.MailID,
		MailType: msl.MailType,
	}
	if out.MailType != enumgame.MailTypeSysMode {
		out.MailInfo = MailInfoFromProto(msl.MailInfo)
	}
	return out
}

func PlayerMailFromProto(msl *pbmail.PlayerMail) *PlayerMail {
	out := &PlayerMail{
		LogID:     msl.LogID,
		SendLogID: msl.SendLogID,
		UserID:    msl.UserID,
		SendID:    msl.SendLogID,
		MailType:  msl.MailType,
	}
	return out
}

func (mi *MailInfo) BeforeCreate(scope *gorm.Scope) error {
	mi.MarshalItemList()
	scope.SetColumn("item_list", mi.ItemList)
	return nil
}

func (msl *MailSendLog) BeforeCreate(scope *gorm.Scope) error {
	msl.MarshalMailStr()
	scope.SetColumn("mail_str", msl.MailStr)
	return nil
}

func (mi *MailInfo) AfterFind() error {
	err := mi.UnmarshalItemList()
	if err != nil {
		return err
	}
	return nil
}

func (msl *MailSendLog) AfterFind() error {
	err := msl.UnmarshalMailStr()
	if err != nil {
		return err
	}
	return nil
}

func (mi *MailInfo) MarshalItemList() error {
	data, _ := json.Marshal(&mi.ItemModeList)
	mi.ItemList = string(data)
	return nil
}

func (mi *MailInfo) UnmarshalItemList() error {
	if len(mi.ItemList) > 0 {
		var out []*ItemModel
		if err := json.Unmarshal([]byte(mi.ItemList), &out); err != nil {
			return err
		}
		mi.ItemModeList = out
	}
	return nil
}

func (msl *MailSendLog) MarshalMailStr() error {
	data, _ := json.Marshal(&msl.MailInfo)
	msl.MailStr = string(data)
	return nil
}

func (msl *MailSendLog) UnmarshalMailStr() error {
	var out *MailInfo
	if err := json.Unmarshal([]byte(msl.MailStr), &out); err != nil {
		return err
	}
	msl.MailInfo = out
	return nil
}
