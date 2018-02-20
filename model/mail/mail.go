package mail

import (
	dbgame "playcards/model/mail/db"
	mdgame "playcards/model/mail/mod"
	mdbill "playcards/model/bill/mod"
	cachegame "playcards/model/mail/cache"
	cacheuser "playcards/model/user/cache"
	enumgame "playcards/model/mail/enum"
	errgame "playcards/model/mail/errors"
	pbgame "playcards/proto/mail"
	mdpage "playcards/model/page"
	mdtime "playcards/model/time"
	enumbill "playcards/model/bill/enum"
	utilproto "playcards/utils/proto"
	"playcards/utils/auth"
	"playcards/model/bill"
	"playcards/utils/log"
	"playcards/utils/db"
	"github.com/jinzhu/gorm"
	"time"
	"fmt"
)

func CreateMailInfo(mi *mdgame.MailInfo) (*mdgame.MailInfo, error) {
	return dbgame.CreateMailInfo(db.DB(), mi)
}

func UpdateMailInfo(mi *mdgame.MailInfo) (*mdgame.MailInfo, error) {
	if mi.MailID < 1 {
		return nil, errgame.ErrMailInfoNotFind
	}
	return dbgame.UpdateMailInfo(db.DB(), mi)
}

func SendSysMail(mailid int32, ids []int32, args []string) error {
	mdMi, err := cachegame.GetMailInfo(mailid)
	if err != nil {
		return err
	}
	if mdMi == nil {
		return errgame.ErrMailInfoNotFind
	}
	content := mdMi.MailContent
	if args != nil && len(args) > 0 {
		argsTemp := make([]interface{}, len(args))
		for i, v := range args {
			argsTemp[i] = v
		}
		content = fmt.Sprintf(mdMi.MailContent, argsTemp...)
	}

	mi := &mdgame.MailInfo{
		MailID:      mdMi.MailID,
		MailName:    mdMi.MailName,
		MailTitle:   mdMi.MailTitle,
		MailContent: content,
		MailType:    mdMi.MailType,
		ItemList:    mdMi.ItemList,
		CreatedAt:   mdMi.CreatedAt,
		UpdatedAt:   mdMi.UpdatedAt,
	}
	msl := &mdgame.MailSendLog{
		MailID:   mi.MailID,
		MailType: enumgame.MailTypeSysMode,
		MailInfo: mi,
	}
	_, err = SendMail(auth.SystemOpUserID, msl, ids, "")
	return err
}

func SendMail(uid int32, msl *mdgame.MailSendLog, ids []int32, channel string) (*mdgame.MailSendLog, error) {
	if len(ids) == 0 && len(channel) == 0 {
		return nil, errgame.ErrSendAndChannel
	}
	if msl.MailID == 0 && msl.MailInfo != nil {
		return nil, errgame.ErrMailInfoContent
	}
	if msl.MailID > 0 {
		msl.MailType = enumgame.MailTypeSysMode
	} else {
		msl.MailType = enumgame.MailTypeSysCustom
	}
	//if msl.MailType == enumgame.MailTypeSysMode && msl.MailID == 0 {
	//	return nil, errgame.ErrMailInfoID
	//}
	if msl.MailType == enumgame.MailTypeSysMode && msl.MailInfo == nil {
		mi, err := cachegame.GetMailInfo(msl.MailID)
		if err != nil {
			return nil, err
		}
		if mi == nil {
			return nil, errgame.ErrMailInfoNotFind
		}
		msl.MailInfo = mi
	}
	if msl.MailInfo == nil {
		return nil, errgame.ErrMailInfoContent
	}
	//if msl.MailInfo.ItemList != "" {
	//	itemInfo := strings.Split(msl.MailInfo.ItemList, ":")
	//	if len(itemInfo) != 4 {
	//		return nil, errgame.ErrItemFormat
	//	}
	//}
	var err error
	msl.SendID = uid
	msl.Status = enumgame.Success
	f := func(tx *gorm.DB) error {
		msl, err = dbgame.CreateMailSendLog(tx, msl)
		if err != nil {
			return err
		}
		return nil
	}
	err = db.Transaction(f)
	if err != nil {
		log.Err("create mail send log failed, %v", err)
		return nil, nil
	}
	go createPlayerMail(msl, ids, channel)
	return msl, err
}

func createPlayerMail(msl *mdgame.MailSendLog, ids []int32, channel string) {
	var haveItem int32 = 0
	if len(msl.MailInfo.ItemModeList) > 0 {
		haveItem = 1
	}
	var pms []*mdgame.PlayerMail
	if len(channel) > 0 {
		ids = dbgame.GetAllUser(db.DB(), channel)
	}
	for _, id := range ids {
		_, u := cacheuser.GetUserByID(id)
		if u == nil {
			continue
		}
		pm := &mdgame.PlayerMail{
			SendLogID: msl.LogID,
			MailID:    msl.MailID,
			UserID:    id,
			SendID:    msl.SendID,
			MailType:  msl.MailType,
			Status:    enumgame.PlayermailNon,
			HaveItem:  haveItem,
		}
		pms = append(pms, pm)
	}
	f := func(tx *gorm.DB) error {
		dbgame.CreatePlayerMails(db.DB(), pms, msl)
		return nil
	}
	err := db.Transaction(f)
	if err != nil {
		log.Err("create player mails failed, %v", err)
	}
}

func ReadMail(uid int32, logid int32) error {
	pm, err := cachegame.GetPlayerMail(uid, logid)
	if err != nil {
		return err
	}
	if pm == nil {
		return errgame.ErrMailNotFind
	}

	if pm.Status > enumgame.PlayermailRead {
		return errgame.ErrMailNotFind
	}
	if pm.HaveItem == 0 {
		err = cachegame.DeletePlayerMail(uid, pm.LogID)
		if err != nil {
			return err
		}
	} else {
		pm.Status = enumgame.PlayermailRead
		err = cachegame.UpdatePlayerMail(pm)
		if err != nil {
			return err
		}
	}
	f := func(tx *gorm.DB) error {
		pm, err = dbgame.UpdatePlayerMail(tx, pm)
		if err != nil {
			return err
		}
		return nil
	}
	err = db.Transaction(f)
	if err != nil {
		return err
	}
	return nil
}

func GetAllMailItems(uid int32) ([]*mdgame.ItemModel, error) {
	pms, err := cachegame.GetPlayerMails(uid)
	if err != nil {
		return nil, err
	}
	if pms == nil {
		return nil, errgame.ErrMailNotNow
	}
	var out []*mdgame.ItemModel
	m := make(map[string]*mdgame.ItemModel)
	for _, pm := range pms {
		if pm.HaveItem == 0 {
			continue
		}
		ils, err := GetMailItems(uid, pm.LogID)
		if err != nil {
			continue
		}
		for _, item := range ils {
			key := fmt.Sprintf("%d-%d-%d", item.MainType, item.SubType, item.ItemID)
			if _, ok := m[key]; ok {
				m[key].Count+=item.Count
			}else{
				m[key] = item
			}
		}
	}
	for _,v := range m{
		out = append(out,v)
	}
	return out, nil
}

func GetMailItems(uid int32, logid int32) ([]*mdgame.ItemModel, error) {
	pm, err := cachegame.GetPlayerMail(uid, logid)
	if err != nil {
		return nil, err
	}
	if pm == nil {
		return nil, errgame.ErrMailNotFind
	}
	if pm.HaveItem == 0 {
		return nil, errgame.ErrHasNotItem
	}

	msl, err := cachegame.GetMailSendLog(pm.SendLogID)
	if err != nil {
		return nil, err
	}
	if msl == nil {
		return nil, errgame.ErrMailNotFind
	}

	if pm.Status > enumgame.PlayermailReadClose {
		return nil, errgame.ErrArealyGetMailItem
	}
	err = cachegame.DeletePlayerMail(uid, pm.LogID)
	if err != nil {
		return nil, err
	}
	f := func(tx *gorm.DB) error {
		pm, err = dbgame.UpdatePlayerMail(tx, pm)
		if err != nil {
			return err
		}
		return nil
	}
	err = db.Transaction(f)
	if err != nil {
		return nil, err
	}
	err = awardItemList(uid,msl)
	if err != nil {
		return nil, err
	}
	return msl.MailInfo.ItemModeList, nil
}

func RefreshAllMailInfoFromDB() error {
	mis, err := dbgame.GetMailInfos(db.DB())
	if err != nil {
		return err
	}

	return cachegame.SetMailInfos(mis)
}

func RefreshAllSendMailLogFromDB() error {
	msl, err := dbgame.GetMailSendLogs(db.DB())
	if err != nil {
		return err
	}
	return cachegame.SetMailSendLogs(msl)
}

func RefreshAllPlayerMailFromDB() error {
	pms, err := dbgame.GetPlayerMails(db.DB())
	if err != nil {
		return err
	}
	return cachegame.SetPlayerMails(pms)
}

func PagePlayerMail(page int32, uid int32) (*pbgame.PagePlayerMailReply, error) {
	var list []*pbgame.PlayerMail
	pms, count, _ := cachegame.GetPlayerMailByID(page, uid)
	//fmt.Printf("PagePlayerMail:%d|%d\n",pms,count)
	mpr := &mdpage.PageReply{
		PageNow:   page,
		PageTotal: count,
	}
	out := &pbgame.PagePlayerMailReply{
		PageReply: mpr.ToProto(),
	}
	if pms == nil && len(pms) == 0 {
		return out, nil
	}
	for _, pbpm := range pms {
		pm := &pbgame.PlayerMail{
			LogID:     pbpm.LogID,
			SendLogID: pbpm.SendLogID,
			UserID:    pbpm.UserID,
			MailType:  pbpm.MailType,
			CreatedAt: mdtime.TimeToProto(pbpm.CreatedAt),
		}
		msl, _ := cachegame.GetMailSendLog(pm.SendLogID)
		pm.MailTitle = msl.MailInfo.MailTitle
		pm.MailContent = msl.MailInfo.MailContent
		//pm.ItemList = msl.MailInfo.ItemModeList
		utilproto.ProtoSlice(msl.MailInfo.ItemModeList, &pm.ItemList)
		_, sn := cacheuser.GetUserByID(pbpm.SendID)
		pm.SendName = sn.Nickname
		list = append(list, pm)
	}
	out.List = list
	return out, nil
}

func CleanOverdueByCreateAt() {
	f := func(tx *gorm.DB) error {
		dbgame.CleanOverdueByCreateAt(tx)
		return nil
	}
	err := db.Transaction(f)
	if err != nil {
		log.Err("clean overdue mail failed, %v", err)
	}
}

func awardItemList(uid int32,sendLog *mdgame.MailSendLog) error {
	//itemInfo := strings.Split(sendLog.MailInfo.ItemList, ":")
	//if len(itemInfo) != 4 {
	//	return errgame.ErrItemFormat
	//}
	//itemType := itemInfo[0]
	//itemSubType := itemInfo[1]
	//itemID := itemInfo[2]
	//itemCount := itemInfo[3]
	for _, item := range sendLog.MailInfo.ItemModeList {
		switch item.MainType {
		case enumgame.CurrencyType:
			currencyType := 0
			if item.SubType == enumgame.CurrencySubTypeGold {
				currencyType = enumbill.TypeGold
			} else if item.SubType == enumgame.CurrencySubTypeDiamond {
				currencyType = enumbill.TypeDiamond
			}
			_, err := bill.GainBalance(uid, time.Now().Unix(), enumbill.JournalTypeMailTitem,
				&mdbill.Balance{Amount: item.Count, CoinType: int32(currencyType)},
			)
			if err != nil {
				return err
			}
			break
		case enumgame.ItemType:
			//TODO 道具物品类型分解
			if item.ItemID == 0 {
			}
			break

		}
	}

	return nil
}

func PageMailInfo(page *mdpage.PageOption, mi *mdgame.MailInfo) (
	[]*mdgame.MailInfo, int64, error) {
	return dbgame.PageMailInfos(db.DB(), page, mi)
}

func PageMailSendLog(page *mdpage.PageOption, msl *mdgame.MailSendLog) (
	[]*mdgame.MailSendLog, int64, error) {
	return dbgame.PageMailSendLogs(db.DB(), page, msl)
}

func PageAllPlayerMail(page *mdpage.PageOption, pm *mdgame.PlayerMail) (
	[]*mdgame.PlayerMail, int64, error) {
	return dbgame.PagePlayerMails(db.DB(), page, pm)
}
