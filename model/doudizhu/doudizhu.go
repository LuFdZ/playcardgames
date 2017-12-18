package doudizhu

import (
	mdroom "playcards/model/room/mod"
	dbroom "playcards/model/room/db"
	enumroom "playcards/model/room/enum"
	cacheroom "playcards/model/room/cache"
	errroom "playcards/model/room/errors"
	mdddz "playcards/model/doudizhu/mod"
	dbddz "playcards/model/doudizhu/db"
	errddz "playcards/model/doudizhu/errors"
	cacheddz "playcards/model/doudizhu/cache"
	enumddz "playcards/model/doudizhu/enum"
	enumbill "playcards/model/bill/enum"
	mdbill "playcards/model/bill/mod"
	pbddz "playcards/proto/doudizhu"
	"playcards/model/bill"
	"encoding/json"
	"playcards/utils/db"
	"time"
	"fmt"
	"playcards/utils/log"
	"github.com/yuin/gopher-lua"
	"github.com/jinzhu/gorm"
	"playcards/utils/tools"
)

var goLua *lua.LState

func InitGoLua(gl *lua.LState) {
	goLua = gl
}

func CreateDoudizhu() []*mdddz.Doudizhu {
	f := func(r *mdroom.Room) bool {
		if r.Status == enumroom.RoomStatusAllReady && r.GameType == enumroom.DoudizhuGameType {
			return true
		}
		return false
	}
	rooms := cacheroom.GetAllRoom(f)
	if rooms == nil && len(rooms) == 0 {
		return nil
	}
	if len(rooms) == 0 {
		return nil
	}
	var newGames []*mdddz.Doudizhu

	for _, room := range rooms {
		var status int32
		status = enumddz.GameStatusInit

		baseScore, gameInit, err := getCardList(room)
		if err != nil {
			continue
		}
		now := gorm.NowFunc()
		doudizhu := &mdddz.Doudizhu{
			RoomID:           room.RoomID,
			Status:           status,
			Index:            room.RoundNow,
			DizhuCardList:    gameInit.DiZhuCardList,
			UserCardInfoList: gameInit.UserCardList,
			GetBankerLogList: gameInit.GetBankerList,
			OpIndex:          0,
			OpTotalIndex:     0,
			SubmitCardNow:    &mdddz.SubmitCard{NextID: gameInit.UserCardList[0].UserID},
			OpID:             gameInit.UserCardList[0].UserID,
			BankerID:         enumddz.DDZCallBanker,
			BaseScore:        baseScore,
			BombScore:        baseScore,
			SubDateAt:        &now,
			OpDateAt:         &now,
			Ids:              room.Ids,
		}

		f := func(tx *gorm.DB) error {
			if room.RoundNow == 1 {
				if room.RoomType != enumroom.RoomTypeClub && room.Cost != 0 {
					err := bill.GainGameBalance(room.PayerID, room.RoomID, enumbill.JournalTypeNiuniu,
						enumbill.JournalTypeNiuniuUnFreeze, &mdbill.Balance{Amount: room.Cost, CoinType: room.CostType})
					if err != nil {
						return err
					}
				}

				for _, user := range room.Users {
					pr := &mdroom.PlayerRoom{
						UserID:    user.UserID,
						RoomID:    room.RoomID,
						GameType:  room.GameType,
						PlayTimes: 0,
					}
					err := dbroom.CreatePlayerRoom(tx, pr)
					if err != nil {
						log.Err("doudizhu create player room err:%v|%v\n", user, err)
						continue
					}
				}
			}
			room.Status = enumroom.RoomStatusStarted
			_, err := dbroom.UpdateRoom(tx, room)
			if err != nil {
				log.Err("Create doudizhu room db err:%v|%v", room, err)
				return err
			}
			err = dbddz.CreateDoudizhu(tx, doudizhu)
			if err != nil {
				log.Err("Create doudizhu db err:%v|%v", doudizhu, err)
				return err
			}
			err = cacheroom.UpdateRoom(room)
			if err != nil {
				log.Err("doudizhu update room set redis failed,%v | %v", room,
					err)
				return err
			}

			err = cacheddz.SetGame(doudizhu, room.Password)
			if err != nil {
				log.Err("doudizhu create set redis failed,%v | %v", doudizhu,
					err)
				return err
			}
			return nil
		}

		err = db.Transaction(f)
		if err != nil {
			log.Err("doudizhu create failed,%v | %v", doudizhu, err)
			continue
		}
		newGames = append(newGames, doudizhu)
	}
	return newGames
}

func getCardList(room *mdroom.Room) (int32, *mdddz.GameInit, error) {
	var roomparam *mdroom.DoudizhuRoomParam
	if err := json.Unmarshal([]byte(room.GameParam), &roomparam); err != nil {
		log.Err("doudizhu unmarshal room param failed, %v", err)
		return 0, nil, errddz.ErrGetCardList
	}
	if room.RoundNow == 1 {
		roomparam.PreBankerIndex = 0
	}
	startIndex := roomparam.PreBankerIndex
	gameInit := &mdddz.GameInit{}
	var userCardList []*mdddz.UserCard
	var getBankerList []*mdddz.GetBanker
	for _, user := range room.Users {
		ui := &mdddz.UserCard{
			UserID:     user.UserID,
			CardList:   []string{},
			CardRemain: []string{},
		}
		userCardList = append(userCardList, ui)
	}
	if startIndex > 0 {
		for i := startIndex; i > 0; i-- {
			for i := int32(0); i < startIndex; i++ {
				tmp := userCardList[i]
				userCardList[i] = userCardList[startIndex]
				userCardList[startIndex] = tmp
			}
		}
	}
	for i, user := range userCardList {
		//var nextID int
		//if i < 4 {
		//	nextID = i + 1
		//} else {
		//	nextID = 0
		//}
		ugb := &mdddz.GetBanker{
			Index:  int32(i),
			UserID: user.UserID,
			Type:   enumddz.DDZCallBanker,
			//NextID:  room.Users[nextID].UserID,
			OpTimes: 0,
		}
		getBankerList = append(getBankerList, ugb)
	}

	gameInit.UserCardList = userCardList
	gameInit.GetBankerList = getBankerList
	if err := goLua.DoString("return G_Reset()"); err != nil {
		log.Err("doudizhu G_Reset %+v", err)
		return 0, nil, errddz.ErrGetCardList
	}

	if err := goLua.DoString(fmt.Sprintf("return G_GetCards()")); err != nil {
		log.Err("doudizhu return logic get cards %v", err)
		return 0, nil, errddz.ErrGetCardList
	}
	getCards := goLua.Get(-1)
	goLua.Pop(1)
	TotalMap := getCards.(*lua.LTable)
	//fmt.Printf()
	var userIndex int32 = 0
	TotalMap.ForEach(func(key lua.LValue, value lua.LValue) {
		CardsList := value.(*lua.LTable)
		CardsList.ForEach(func(key lua.LValue, value lua.LValue) {
			if userIndex == 4 {
				gameInit.DiZhuCardList = append(gameInit.DiZhuCardList, value.String())
			} else {
				userCardList[userIndex].CardList = append(userCardList[userIndex].CardList, value.String())
				userCardList[userIndex].CardRemain = userCardList[userIndex].CardList
			}
		})
		userIndex++
	})
	return roomparam.BaseScore, gameInit, nil
}

func updateGame(ddz *mdddz.Doudizhu, toDB bool) error {
	err := cacheddz.UpdateGame(ddz)
	if err != nil {
		log.Err("doudizhu create set redis failed,%v | %v", ddz,
			err)
		return err
	}
	if toDB {
		f := func(tx *gorm.DB) error {
			_, err := dbddz.UpdateDoudizhu(tx, ddz)
			if err != nil {
				log.Err("Create doudizhu db err:%v|%v", ddz, err)
				return err
			}
			return nil
		}
		err = db.Transaction(f)
		if err != nil {
			return err
		}
	}
	return nil
}

func UpdateGame() []*mdddz.Doudizhu {
	f := func(r *mdddz.Doudizhu) bool {
		if r.Status < enumddz.GameStatusDone {
			return true
		}
		return false
	}
	ddzs := cacheddz.GetAllDDZ(f)
	if len(ddzs) == 0 {
		return nil
	}
	var updateGames []*mdddz.Doudizhu
	for _, ddz := range ddzs {
		if ddz.Status == enumddz.GameStatusInit {
			sub := time.Now().Sub(*ddz.OpDateAt)
			if sub.Seconds() < enumddz.GetBankerCountDown {
				continue
			}
		} else if ddz.Status == enumddz.GameStatusSubmitCard {
			sub := time.Now().Sub(*ddz.OpDateAt)
			if sub.Seconds() < enumddz.SubmitCardCountDown {
				continue
			}
		} else if ddz.Status == enumddz.GameStatusStarted {
			pwd := cacheddz.GetRoomPaawordRoomID(ddz.RoomID)
			room, err := cacheroom.GetRoom(pwd)
			if err != nil {
				log.Err("doudizhu room get session failed, roomid:%d,pwd:%s,err:%v", ddz.RoomID, pwd, err)
				continue
			}
			if room == nil {
				log.Err("doudizhu room get session nil, %v|%d", pwd, ddz.RoomID)
				continue
			}

			var roomparam *mdroom.NiuniuRoomParam
			if err := json.Unmarshal([]byte(room.GameParam),
				&roomparam); err != nil {
				log.Err("doudizhu unmarshal room param failed, %v",
					err)
				continue
			}

			data, _ := json.Marshal(&roomparam)
			room.GameParam = string(data)
			//niuniu.Status = enumniu.GameStatusDone
			//niuniu.BroStatus = enumniu.GameStatusDone
			//room.Status = enumr.RoomStatusReInit
			//
			//f := func(tx *gorm.DB) error {
			//	niuniu, err = dbniu.UpdateNiuniu(tx, niuniu)
			//	if err != nil {
			//		log.Err("niuniu update db failed, %v|%v", niuniu, err)
			//		return err
			//	}
			//	room, err = dbr.UpdateRoom(tx, room)
			//	if err != nil {
			//		log.Err("niuniu update room db failed, %v|%v", niuniu, err)
			//		return err
			//	}
			//	return nil
			//}
			//err = db.Transaction(f)
			//if err != nil {
			//	//print(err)
			//	log.Err("niuniu update failed, %v", err)
			//	continue
			//}
			//err = cacheniu.DeleteGame(niuniu.RoomID)
			//if err != nil {
			//	log.Err("niuniu room del session failed, roomid:%d,pwd:%s,err:%v", niuniu.RoomID, pwd, err)
			//	continue
			//}
			//
			//err = cacher.UpdateRoom(room)
			//if err != nil {
			//	log.Err("niuniu room update room redis failed,%v | %v",
			//		room, err)
			//	continue
			//}

		}
		updateGames = append(updateGames, ddz)
	}

	return updateGames
}

func GetBanker(uid int32, gid int32, getBanker int32, room *mdroom.Room) (int32, *mdddz.Doudizhu, error) {
	if getBanker < 1 || getBanker > 3 {
		return 0, nil, errddz.ErrGetBankerParam
	}
	if room.Status > enumroom.RoomStatusStarted {
		return 0, nil, errroom.ErrGameIsDone
	}
	if room.Giveup == enumroom.WaitGiveUp {
		return 0, nil, errroom.ErrInGiveUp
	}
	ddz, err := cacheddz.GetGame(room.RoomID)
	if err != nil {
		return 0, nil, err
	}
	if ddz == nil {
		return 0, nil, errddz.ErrGameNotExist
	}
	if ddz.GameID != gid {
		return 0, nil, errddz.ErrGameDiffer
	}
	if ddz.Status != enumddz.GameStatusInit {
		return 0, nil, errddz.ErrBankerDone
	}
	if ddz.OpID != uid {
		return 0, nil, errddz.ErrNotYouTrun
	}
	if getBanker != ddz.BankerType {
		return 0, nil, errddz.ErrBankerType
	}
	if ddz.BankerType == enumddz.DDZBankerTypeCall && getBanker == enumddz.DDZBankerTypeGet {
		ddz.BankerType = enumddz.DDZBankerTypeGet
	}
	ugb, bankerStatus := getNextBanker(ddz.OpTotalIndex, ddz.OpIndex, getBanker, ddz.GetBankerLogList)
	ugb.OpTimes++
	ddz.OpTotalIndex++
	strLog := fmt.Sprintf("%d|%d|%d|%d", ddz.OpTotalIndex, uid, getBanker, bankerStatus)
	ddz.GetBankerLogStrList = append(ddz.GetBankerLogStrList, strLog)
	toDB := false
	switch bankerStatus {
	case enumddz.DDZBankerStatusReStart:
		_, gameInit, err := getCardList(room)
		if err != nil {
			return 0, nil, err
		}
		ddz.OpTotalIndex = 0
		ddz.DizhuCardList = gameInit.DiZhuCardList
		ddz.UserCardInfoList = gameInit.UserCardList
		ddz.OpIndex = 0
		ddz.OpID = gameInit.UserCardList[0].UserID
		ddz.BankerID = enumddz.DDZCallBanker
		ddz.RestartTimes++
		toDB = true
		break
	case enumddz.DDZBankerStatusContinue:

		ddz.OpIndex = ugb.Index
		ddz.OpID = ugb.UserID
		break
	case enumddz.DDZBankerStatusFinish:
		ddz.OpTotalIndex = 0
		ddz.OpIndex = ugb.Index
		ddz.OpID = ugb.UserID
		ddz.BankerID = ugb.UserID
		ddz.Status = enumddz.GameStatusSubmitCard
		cardList := ddz.UserCardInfoList[ugb.Index]
		cardList.CardList = tools.MergeSlice(cardList.CardList, ddz.DizhuCardList)
		break
	}
	ddz.BombScore += ddz.BaseScore
	now := gorm.NowFunc()
	ddz.OpDateAt = &now
	ddz.SubDateAt = &now
	updateGame(ddz, toDB)
	return bankerStatus, ddz, nil
}

func SubmitCard(uid int32, gid int32, submitCardList []string, room *mdroom.Room) ([]string, *pbddz.SubmitCard, *mdddz.Doudizhu, error) {
	if room.Status > enumroom.RoomStatusStarted {
		return nil, nil, nil, errroom.ErrGameIsDone
	}
	if room.Giveup == enumroom.WaitGiveUp {
		return nil, nil, nil, errroom.ErrInGiveUp
	}
	game, err := cacheddz.GetGame(room.RoomID)
	if err != nil {
		return nil, nil, nil, err
	}
	if game.GameID != gid {
		return nil, nil, nil, errddz.ErrGameDiffer
	}
	if game.Status != enumddz.GameStatusSubmitCard {
		return nil, nil, nil, errddz.ErrSubmitCardDone
	}
	if game.OpID != uid {
		return nil, nil, nil, errddz.ErrNotYouTrun
	}
	uci := game.UserCardInfoList[game.OpIndex]
	submitCardNumber := len(submitCardList)
	isSelf := false
	if game.SubmitCardNow.UserID == uid {
		isSelf = true
	}
	var cardType int32
	if submitCardNumber > 0 {
		checkHasCard := checkHasCards(submitCardList, uci.CardRemain)
		if !checkHasCard {
			return nil, nil, nil, errddz.ErrCardNotExist
		}
		result, bombScore, cType := cardCompare(game.SubmitCardNow, submitCardList, isSelf)
		cardType = cType
		if !result {
			return nil, nil, nil, errddz.ErrSubmitCard
		}
		game.BombScore *= bombScore
		if isSelf {
			for _, u := range game.UserCardInfoList {
				u.CardList = nil
			}
		}
		uci.CardLast = submitCardList
		game.SubmitCardNow.CardList = submitCardList
		game.SubmitCardNow.CardType = cardType
		game.SubmitCardNow.UserID = uid
		deleteCard(uci, submitCardList)
		if len(uci.CardRemain) == 0 {
			game.WinerID = uid
			game.Status = enumddz.GameStatusStarted
			if uid == game.BankerID {
				game.WinerType = enumddz.Dizhu
			} else {
				game.WinerType = enumddz.NongMin
			}
		}
	} else {
		if isSelf {
			return nil, nil, nil, errddz.ErrSubmitCard
		} else {
			game.SubmitCardNow.UserID = uid
		}
	}

	game.OpTotalIndex++
	nextIndex := getNetxIndex(game.OpIndex)
	game.SubmitCardNow.NextID = game.UserCardInfoList[nextIndex].UserID
	now := gorm.NowFunc()
	game.OpDateAt = &now
	game.SubDateAt = &now

	submitCardLogStr := fmt.Sprintf("%d|%d|%d|%v", game.OpTotalIndex, uid, game.BombScore, submitCardList)
	game.GameCardLogList = append(game.GameCardLogList, submitCardLogStr)
	err = cacheddz.UpdateGame(game)
	if err != nil {
		log.Err("doudizhu set redis failed, %v", err)
		return nil, nil, nil, err
	}

	msg := &pbddz.SubmitCard{
		GameID:    gid,
		SubmitID:  uid,
		CardType:  cardType,
		NextID:    game.SubmitCardNow.NextID,
		BombScore: game.BombScore,
		Status:    game.Status,
		CountDown: &pbddz.CountDown{
			ServerTime: game.OpDateAt.Unix(),
			Count:      enumddz.SubmitCardCountDown,
		},
		CardList: submitCardList,
		//CardRemain:       uci.CardRemain,
		CardRemainNumber: int32(len(uci.CardRemain)),
	}

	if game.Status == enumddz.GameStatusStarted {
		err = GetGameResult(game, room)

	} else {
		err = updateGame(game, false)
	}
	if err != nil {
		return nil, nil, nil, err
	}

	return uci.CardRemain, msg, game, nil //ddz, nil
}

func GetGameResult(game *mdddz.Doudizhu, room *mdroom.Room) error {
	var score = game.BombScore
	for _, uc := range game.UserCardInfoList {
		ur := &mdddz.UserResult{}
		ur.UserID = uc.UserID
		if game.WinerType == enumddz.Dizhu {
			if ur.UserID == game.BankerID {
				ur.Score = score
			} else {
				ur.Score = -score
			}
		} else {
			if ur.UserID == game.BankerID {
				ur.Score = -score
			} else {
				ur.Score = score
			}
		}
	}
	game.Status = enumddz.GameStatusDone
	f := func(tx *gorm.DB) error {
		game, err := dbddz.UpdateDoudizhu(tx, game)
		if err != nil {
			log.Err("doudizhu update db failed, %v|%v", game, err)
			return err
		}
		room, err = dbroom.UpdateRoom(tx, room)
		if err != nil {
			log.Err("doudizhu update room db failed, %v|%v", room, err)
			return err
		}
		return nil
	}
	err := db.Transaction(f)
	if err != nil {
		//print(err)
		log.Err("doudizhu update failed, %v", err)
		return err
	}
	err = cacheddz.DeleteGame(game.RoomID)
	if err != nil {
		log.Err("doudizhu room del session failed, roomid:%d,pwd:%s,err:%v", game.RoomID, room.Password, err)
		return err
	}

	err = cacheroom.UpdateRoom(room)
	if err != nil {
		log.Err("doudizhu room update room redis failed,%v | %v",
			room, err)
		return err
	}
	return nil
}

func getNetxIndex(startIndex int32) int32 {
	if startIndex == 3 {
		return 0
	}
	return startIndex + 1
}

func cardCompare(SubmitCardNow *mdddz.SubmitCard, submitCardList []string, isSelf bool) (bool, int32, int32) {
	return true, 0, 0
}

func checkHasCards(submitCards []string, cardList []string) bool {

	for i := 0; i < len(submitCards); i++ {
		for j := 0; j < len(cardList); j++ {
			if submitCards[i] == cardList[j] {
				submitCards[i] = "ok"
				cardList[j] = "pass"
				continue
			}
		}
	}

	for _, sc := range submitCards {
		if len(sc) > 2 {
			return false
		}
	}
	return true
}

func deleteCard(ui *mdddz.UserCard, submitCardList []string) {
	var newCardList []string
	for i := 0; i < len(submitCardList); i++ {
		for j := 0; j < len(ui.CardRemain); j++ {
			if submitCardList[i] != ui.CardRemain[j] {
				newCardList = append(newCardList, ui.CardList[j])
			}
		}
	}
	ui.CardRemain = newCardList
}

func getNextBanker(opIndex int32, startIndex int32, bankerType int32, ugbList []*mdddz.GetBanker) (*mdddz.GetBanker, int32) {
	fmt.Printf("getNextBanker:%d|%d|%+v\n", opIndex, startIndex, ugbList)
	ugbList[startIndex].Type = bankerType
	if opIndex < 3 {
		return ugbList[startIndex+1], enumddz.DDZBankerStatusContinue
	} else {
		var noCallCount int32 = 0
		lastCallIndex := -1
		lastGetIndex := 0
		for i := 0; i < 4; i++ {
			if ugbList[i].Type == enumddz.DDZCallBanker {
				lastCallIndex = i
			} else if ugbList[i].Type == enumddz.DDZGetBanker {
				lastGetIndex = i
			} else if ugbList[i].Type == enumddz.DDZNoBanker {
				noCallCount++
			}
		}
		if noCallCount == 4 {
			return nil, enumddz.DDZBankerStatusReStart
		} else if lastCallIndex > -1 {
			return ugbList[lastCallIndex], enumddz.DDZBankerStatusContinue
		} else {
			return ugbList[lastGetIndex], enumddz.DDZBankerStatusFinish
		}
	}
}
