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
	"playcards/model/room"
	"encoding/json"
	"playcards/utils/db"
	"time"
	"fmt"
	"playcards/utils/log"
	"github.com/yuin/gopher-lua"
	"github.com/jinzhu/gorm"
	"playcards/utils/tools"
	"strings"
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
	rooms := cacheroom.GetAllRooms(f)
	if rooms == nil && len(rooms) == 0 {
		return nil
	}
	if len(rooms) == 0 {
		return nil
	}
	var newGames []*mdddz.Doudizhu

	for _, mdr := range rooms {
		var status int32
		status = enumddz.GameStatusInit

		baseScore, gameInit, err := getCardList(mdr)
		if err != nil {
			continue
		}
		now := gorm.NowFunc()
		doudizhu := &mdddz.Doudizhu{
			RoomID:           mdr.RoomID,
			Status:           status,
			Index:            mdr.RoundNow,
			DizhuCardList:    gameInit.DiZhuCardList,
			UserCardInfoList: gameInit.UserCardList,
			GetBankerLogList: gameInit.GetBankerList,
			OpIndex:          0,
			OpTotalIndex:     0,
			SubmitCardNow:    &mdddz.SubmitCard{NextID: gameInit.UserCardList[0].UserID},
			OpID:             gameInit.UserCardList[0].UserID,
			BankerID:         0,
			BaseScore:        baseScore,
			BombTimes:        1,
			BankerTimes:      1,
			PassWord:         mdr.Password,
			SubDateAt:        &now,
			OpDateAt:         &now,
			Ids:              mdr.Ids,
		}

		f := func(tx *gorm.DB) error {
			if mdr.RoundNow == 1 {
				if mdr.RoomType != enumroom.RoomTypeClub && mdr.Cost != 0 {
					err := bill.GainGameBalance(mdr.PayerID, mdr.RoomID, enumbill.JournalTypeNiuniu,
						enumbill.JournalTypeNiuniuUnFreeze, &mdbill.Balance{Amount: mdr.Cost, CoinType: mdr.CostType})
					if err != nil {
						return err
					}
				}

				for _, user := range mdr.Users {
					pr := &mdroom.PlayerRoom{
						UserID:    user.UserID,
						RoomID:    mdr.RoomID,
						GameType:  mdr.GameType,
						PlayTimes: 0,
					}
					err := dbroom.CreatePlayerRoom(tx, pr)
					if err != nil {
						log.Err("doudizhu create player room err:%v|%v\n", user, err)
						continue
					}
				}
			}
			mdr.Status = enumroom.RoomStatusStarted
			_, err := dbroom.UpdateRoom(tx, mdr)
			if err != nil {
				log.Err("Create doudizhu room db err:%v|%v", mdr, err)
				return err
			}
			err = dbddz.CreateDoudizhu(tx, doudizhu)
			if err != nil {
				log.Err("Create doudizhu db err:%v|%v", doudizhu, err)
				return err
			}
			err = cacheroom.UpdateRoom(mdr)
			if err != nil {
				log.Err("doudizhu update room set redis failed,%v | %v", mdr,
					err)
				return err
			}

			err = cacheddz.SetGame(doudizhu)
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

func getCardList(mdr *mdroom.Room) (int32, *mdddz.GameInit, error) {
	var roomparam *mdroom.DoudizhuRoomParam
	if err := json.Unmarshal([]byte(mdr.GameParam), &roomparam); err != nil {
		log.Err("doudizhu unmarshal room param failed, %v", err)
		return 0, nil, errddz.ErrGetCardList
	}
	if mdr.RoundNow == 1 {
		roomparam.PreBankerIndex = 0
	}
	startIndex := roomparam.PreBankerIndex
	gameInit := &mdddz.GameInit{}
	var userCardList []*mdddz.UserCard
	var getBankerList []*mdddz.GetBanker
	for _, user := range mdr.Users {
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
		ugb := &mdddz.GetBanker{
			Index:   int32(i),
			UserID:  user.UserID,
			Type:    enumddz.DDZCallBanker,
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

func updateGame(game *mdddz.Doudizhu, toDB bool) error {
	err := cacheddz.UpdateGame(game)
	if err != nil {
		log.Err("doudizhu create set redis failed,%v | %v", game,
			err)
		return err
	}
	if toDB {
		f := func(tx *gorm.DB) error {
			_, err := dbddz.UpdateDoudizhu(tx, game)
			if err != nil {
				log.Err("Create doudizhu db err:%v|%v", game, err)
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
	ddzs, err := cacheddz.GetAllDoudizhuByStatus(enumddz.GameStatusDone)
	if err != nil {
		log.Err("UpdateGame_GetAllDoudizhuByStatus_Err err:%v", err)
		return nil
	}
	if len(ddzs) == 0 {
		return nil
	}
	var updateGames []*mdddz.Doudizhu
	for _, game := range ddzs {
		if game.Status == enumddz.GameStatusInit {
			sub := time.Now().Sub(*game.OpDateAt)
			if sub.Seconds() < enumddz.GetBankerCountDown || game.GoldGame == 0 {
				continue
			}
		} else if game.Status == enumddz.GameStatusSubmitCard {
			sub := time.Now().Sub(*game.OpDateAt)
			if sub.Seconds() < enumddz.SubmitCardCountDown && game.GoldGame > 0 {
				continue
			}
		} else if game.Status == enumddz.GameStatusStarted {
			mdr, err := cacheroom.GetRoom(game.PassWord)
			if err != nil {
				//print(err)
				log.Err("ddz room get session failed, roomid:%d,pwd:%s,err:%v", game.RoomID, game.PassWord, err)
				continue
			}
			if mdr == nil {
				log.Err("ddz room get session nil, %v|%d", game.PassWord, mdr.RoomID)
				continue
			}
			err = getGameResult(game, mdr)
			if err != nil {
				log.Err("Update_GameGetGameResult_Err gameID:%d,err:%v", game.GameID, err)
				continue
			}
		}
		updateGames = append(updateGames, game)
	}
	return updateGames
}

func GetBanker(uid int32, gid int32, getBanker int32, mdr *mdroom.Room) (*mdddz.Doudizhu, error) {
	if getBanker < 1 || getBanker > 3 {
		return nil, errddz.ErrGetBankerParam
	}
	if mdr.Status > enumroom.RoomStatusStarted {
		return nil, errroom.ErrGameIsDone
	}
	if mdr.Giveup == enumroom.WaitGiveUp {
		return nil, errroom.ErrInGiveUp
	}
	game, err := cacheddz.GetGame(mdr.RoomID)
	if err != nil {
		return nil, err
	}
	if game == nil {
		return nil, errddz.ErrGameNotExist
	}
	if game.GameID != gid {
		return nil, errddz.ErrGameDiffer
	}
	if game.Status != enumddz.GameStatusInit {
		return nil, errddz.ErrBankerDone
	}
	if game.OpID != uid {
		return nil, errddz.ErrNotYouTrun
	}
	if game.BankerType == 0 && (getBanker != enumddz.DDZCallBanker && getBanker != enumddz.DDZNoBanker) {
		return nil, errddz.ErrBankerType
	} else if (game.BankerType == enumddz.DDZCallBanker || game.BankerType == enumddz.DDZGetBanker) &&
		getBanker < enumddz.DDZGetBanker {
		return nil, errddz.ErrBankerType
	}
	if game.BankerType == 0 {
		game.BankerType = enumddz.DDZBankerTypeCall
	} else if game.BankerType == enumddz.DDZBankerTypeCall && getBanker == enumddz.DDZBankerTypeGet {
		game.BankerType = enumddz.DDZBankerTypeGet
	}
	ugb, bankerStatus := getNextBanker(game.OpTotalIndex, game.OpIndex, getBanker, game.GetBankerLogList)
	ugb.OpTimes++
	game.OpTotalIndex++
	strLog := fmt.Sprintf("%d|%d|%d|%d", game.OpTotalIndex, uid, getBanker, bankerStatus)
	game.GetBankerLogStrList = append(game.GetBankerLogStrList, strLog)
	toDB := false
	game.BankerStatus = bankerStatus
	switch game.BankerStatus {
	case enumddz.DDZBankerStatusReStart:
		_, gameInit, err := getCardList(mdr)
		if err != nil {
			return nil, err
		}
		game.OpTotalIndex = 0
		game.DizhuCardList = gameInit.DiZhuCardList
		game.UserCardInfoList = gameInit.UserCardList
		game.OpIndex = 0
		game.OpID = gameInit.UserCardList[0].UserID
		game.BankerID = enumddz.DDZCallBanker
		game.RestartTimes++
		toDB = true
		break
	case enumddz.DDZBankerStatusContinue:

		game.OpIndex = ugb.Index
		game.OpID = ugb.UserID
		break
	case enumddz.DDZBankerStatusFinish:
		game.OpTotalIndex = 0
		game.OpIndex = ugb.Index
		game.OpID = ugb.UserID
		game.BankerID = ugb.UserID
		game.Status = enumddz.GameStatusSubmitCard
		cardList := game.UserCardInfoList[ugb.Index]
		cardList.CardList = tools.MergeSlice(cardList.CardList, game.DizhuCardList)
		break
	}
	game.BankerTimes *= 2
	now := gorm.NowFunc()
	game.OpDateAt = &now
	game.SubDateAt = &now
	updateGame(game, toDB)
	return game, nil
}

func SubmitCard(uid int32, gid int32, submitCardList []string, mdr *mdroom.Room) (*mdddz.Doudizhu, error) {
	if mdr.Status > enumroom.RoomStatusStarted {
		return nil, errroom.ErrGameIsDone
	}
	if mdr.Giveup == enumroom.WaitGiveUp {
		return nil, errroom.ErrInGiveUp
	}
	game, err := cacheddz.GetGame(mdr.RoomID)
	if err != nil {
		return nil, err
	}
	if game.GameID != gid {
		return nil, errddz.ErrGameDiffer
	}
	if game.Status != enumddz.GameStatusSubmitCard {
		return nil, errddz.ErrSubmitCardDone
	}
	if game.OpID != uid {
		return nil, errddz.ErrNotYouTrun
	}
	uci := game.UserCardInfoList[game.OpIndex]
	submitCardNumber := len(submitCardList)
	isSelf := false
	if game.SubmitCardNow.UserID == uid {
		isSelf = true
	}
	var cardType int32
	if submitCardNumber > 0 {
		if cacheroom.GetRoomTestConfigKey("DoudizhuCheckHasCards") != "0" {
			checkHasCard := checkHasCards(submitCardList, uci.CardRemain)
			if !checkHasCard {
				return nil, errddz.ErrCardNotExist
			}
		}
		result, bombScore, cType := cardCompare(game.SubmitCardNow, submitCardList, isSelf)
		cardType = cType
		if !result {
			return nil, errddz.ErrSubmitCard
		}
		game.BombTimes *= bombScore
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
			if game.OpTotalIndex < 3 {
				game.Spring = 2
			}
		}
		bombType, ok := enumddz.BombScoreMap[cType]
		if !ok {
			bombType = 0
		}
		switch bombType {
		case 0:
			game.CommonBomb++
			break
		case 2:

			break
		}
	} else {
		if isSelf {
			return nil, errddz.ErrSubmitCard
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

	submitCardLogStr := fmt.Sprintf("%d|%d|%d|%v", game.OpTotalIndex, uid, game.BombTimes, submitCardList)
	game.GameCardLogList = append(game.GameCardLogList, submitCardLogStr)
	err = cacheddz.UpdateGame(game)
	if err != nil {
		log.Err("doudizhu set redis failed, %v", err)
		return nil, err
	}

	if game.Status != enumddz.GameStatusStarted {
		err = updateGame(game, false)
		game.OpID = game.SubmitCardNow.NextID
		game.OpIndex = nextIndex
		if err != nil {
			return nil, err
		}
	}
	return game, nil //ddz, nil
}

func getGameResult(game *mdddz.Doudizhu, mdr *mdroom.Room) error {
	var score = game.BankerTimes * game.BombTimes * game.BaseScore
	for _, uc := range game.UserCardInfoList {
		ur := &mdddz.UserResult{}
		ur.UserID = uc.UserID
		if game.WinerType == enumddz.Dizhu {
			if ur.UserID == game.BankerID {
				ur.Score = 3 * score
			} else {
				ur.Score = -3 * score
			}
		} else {
			if ur.UserID == game.BankerID {
				ur.Score = -score
			} else {
				ur.Score = score
			}
		}

		for _, userResult := range mdr.UserResults {
			if userResult.UserID == uc.UserID {
				userResult.Score += ur.Score
				if ur.Score > 0 {
					userResult.Win += 1
				} else if ur.Score == 0 {
					userResult.Tie += 1
				} else if ur.Score < 0 {
					userResult.Lost += 1
				}
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
		mdr, err = dbroom.UpdateRoom(tx, mdr)
		if err != nil {
			log.Err("doudizhu update room db failed, %v|%v", mdr, err)
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
	err = cacheddz.DeleteGame(game)
	if err != nil {
		log.Err("doudizhu room del session failed, roomid:%d,pwd:%s,err:%v", game.RoomID, mdr.Password, err)
		return err
	}

	err = cacheroom.UpdateRoom(mdr)
	if err != nil {
		log.Err("doudizhu room update room redis failed,%v | %v",
			mdr, err)
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
			return ugbList[3], enumddz.DDZBankerStatusReStart
		} else if lastCallIndex > -1 {
			return ugbList[lastCallIndex], enumddz.DDZBankerStatusContinue
		} else {
			return ugbList[lastGetIndex], enumddz.DDZBankerStatusFinish
		}
	}
}

func GameResultList(rid int32) (*pbddz.GameResultListReply, error) {
	var list []*pbddz.GameResult
	games, err := dbddz.GetDoudizhuByRoomID(db.DB(), rid)
	if err != nil {
		return nil, err
	}
	for _, game := range games {
		gr := &pbddz.GameResult{
			GameID:    game.GameID,
			BaseScore: game.BaseScore,
			BombTimes: game.BankerTimes * game.BombTimes,
		}
		var userList []*pbddz.UserInfo
		for _, ui := range game.GameResultList {
			userList = append(userList, ui.ToProto())
		}
		gr.UserResult = userList
		list = append(list, gr)
	}
	out := &pbddz.GameResultListReply{
		List: list,
	}
	return out, nil
}

func DoudizhuExist(uid int32, rid int32) (*pbddz.DoudizhuRecoveryReply, error) {
	out := &pbddz.DoudizhuRecoveryReply{}
	_, roomRecovery, err := room.CheckRoomExist(uid, rid)
	if err != nil {
		return nil, err
	}
	out.RoomExist = roomRecovery.ToProto()
	out.RoomExist.Room.CreateOrEnter = enumroom.EnterRoom
	if roomRecovery.Status < enumroom.RecoveryGameStart && roomRecovery.Status != enumroom.RecoveryInitNoReady {
		return out, nil
	}
	game, err := cacheddz.GetGame(roomRecovery.Room.RoomID)
	if err != nil {
		return nil, err
	}
	if game == nil {
		game, err = dbddz.GetLastDoudizhuByRoomID(db.DB(), roomRecovery.Room.RoomID)
		if err != nil {
			return nil, err
		}
		if game == nil {
			return nil, errddz.ErrGameNotExist
		}
	}
	if game.Status == enumddz.GameStatusGiveUp {
		return out, nil
	}
	out.DoudizhuExist.GameID = game.GameID
	out.DoudizhuExist.RoomID = game.RoomID
	out.DoudizhuExist.OpID = game.OpID
	out.DoudizhuExist.Status = enumddz.GameStatusInit
	out.DoudizhuExist.UserCardList = game.GetUserCard(uid).CardList
	out.DoudizhuExist.CountDown = &pbddz.CountDown{game.OpDateAt.Unix(), enumddz.GetBankerCountDown}
	out.DoudizhuExist.DizhuCardList = game.DizhuCardList
	if game.Status >= enumddz.GameStatusInit {
		getBankerTimes := len(game.GetBankerLogStrList)
		if getBankerTimes > 0 {
			lastGetBanker := game.GetBankerLogStrList[getBankerTimes-1]
			getBankerInfo := strings.Split(lastGetBanker, "|")
			//fmt.Sprintf("%d|%d|%d|%d", game.OpTotalIndex, uid, getBanker, bankerStatus)
			out.DoudizhuExist.LastGetBankerID, _ = tools.StringToInt(getBankerInfo[1])
			out.DoudizhuExist.LastGetBankerType, _ = tools.StringToInt(getBankerInfo[2])
		}
	}

	if game.Status >= enumddz.GameStatusSubmitCard {
		submitCardTimes := len(game.GameCardLogList)
		if submitCardTimes > 0 {
			var limit int32 = 4
			for i := submitCardTimes; i > 0 && limit > 0; i-- {
				limit--
				submitCardInfo := strings.Split(game.GameCardLogList[i-1], "|")
				userID, _ := tools.StringToInt(submitCardInfo[1])
				cardList := strings.Split(submitCardInfo[3], ",")
				uci := game.GetUserCard(userID)
				pbui := &pbddz.UserInfo{
					UserID:           userID,
					LastCard:         cardList,
					CardRemainNumber: int32(len(uci.CardRemain)),
				}
				if userID == uid {
					pbui.CardRemain = uci.CardRemain
				}
				out.DoudizhuExist.UserInfoList = append(out.DoudizhuExist.UserInfoList, pbui)
			}
		}
	}
	if game.Status >= enumddz.GameStatusDone {
		for _, ur := range game.GameResultList {
			for _, ui := range out.DoudizhuExist.UserInfoList {
				ui.Score = ur.Score
				ui.CardRemain = game.GetUserCard(ui.UserID).CardRemain
			}
		}
	}
	return out, nil
}

func CleanGame() error {
	var gids []int32
	rids, err := cacheroom.GetAllDeleteRoomKey(enumroom.DoudizhuGameType)
	if err != nil {
		log.Err("get niuniu clean room err:%v", err)
		return err
	}
	for _, rid := range rids {
		game, err := cacheddz.GetGame(rid)
		if err != nil {
			log.Err("get niuniu give up room err:%d|%v", rid, err)
			continue
		}
		if game != nil {
			log.Debug("clean niuniu game:%d|%d|%+v\n", game.GameID, game.RoomID, game.Ids)
			gids = append(gids, game.GameID)
			err = cacheddz.DeleteGame(game)
			if err != nil {
				log.Err(" delete niuniu set session failed, %v",
					err)
				continue
			}
			rid = game.RoomID
		}
		err = cacheroom.CleanDeleteRoom(enumroom.DoudizhuGameType, rid)
		if err != nil {
			log.Err(" delete null game niuniu delete room session failed,roomid:%d,err:%v", rid,
				err)
			continue
		}
	}
	if len(gids) > 0 {
		f := func(tx *gorm.DB) error {
			err = dbddz.GiveUpGameUpdate(tx, gids)
			if err != nil {
				return err
			}
			return nil
		}
		go db.Transaction(f)
	}

	return nil
}
