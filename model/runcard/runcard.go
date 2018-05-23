package runcard

import (
	"encoding/json"
	"fmt"
	cacheroom "playcards/model/room/cache"
	dbroom "playcards/model/room/db"
	enumroom "playcards/model/room/enum"
	errroom"playcards/model/room/errors"
	mdroom "playcards/model/room/mod"
	cachegame "playcards/model/runcard/cache"
	dbgame "playcards/model/runcard/db"
	enumgame "playcards/model/runcard/enum"
	errgame "playcards/model/runcard/errors"
	mdgame "playcards/model/runcard/mod"
	pbgame "playcards/proto/runcard"
	cachelog "playcards/model/log/cache"
	"playcards/model/room"
	"playcards/utils/db"
	"playcards/utils/log"
	"github.com/jinzhu/gorm"
	"github.com/yuin/gopher-lua"
	"time"
	"sort"
	"playcards/utils/tools"
)

func CreateGame(goLua *lua.LState) []*mdgame.Runcard {
	rooms := cacheroom.GetAllRoomByGameTypeAndStatus(enumroom.RunCardGameType, enumroom.RoomStatusAllReady)
	if rooms == nil && len(rooms) == 0 {
		return nil
	}
	if len(rooms) == 0 {
		return nil
	}
	var newGames []*mdgame.Runcard

	for _, mdr := range rooms {
		if mdr.RoundNow == 1 {
			var userResults []*mdroom.GameUserResult
			for _, ur := range mdr.Users {
				userResult := &mdroom.GameUserResult{
					UserID: ur.UserID,
					Role:   ur.Role,
					Win:    0,
					Lost:   0,
					Tie:    0,
					Score:  0,
				}
				userResults = append(userResults, userResult)
			}
			mdr.UserResults = userResults
		} else if len(mdr.Users) != len(mdr.UserResults) {
			for _, ur := range mdr.UserResults {
				hasInResult := false
				for _, u := range mdr.Users {
					if u.UserID == ur.UserID {
						hasInResult = true
					}
				}
				if !hasInResult {
					userResult := &mdroom.GameUserResult{
						UserID: ur.UserID,
						Role:   ur.Role,
						Win:    0,
						Lost:   0,
						Tie:    0,
						Score:  0,
					}
					mdr.UserResults = append(mdr.UserResults, userResult)
				}
			}
		}
		var num int32 = 0
		var uis []*mdgame.UserInfo
		for _, ur := range mdr.Users {
			if ur.UserRole == enumroom.UserRolePlayerBro {
				ui := &mdgame.UserInfo{
					UserID:        ur.UserID,
					Status:        enumgame.UserStatusInit,
					Score:         0,
					BoomNum:       0,
					BoomScore:     0,
					ClubCoinScore: 0,
					HasSubmitCard: 0,
				}
				uis = append(uis, ui)
				num ++
			}
		}

		t := time.Now()
		gr := &mdgame.GameResult{
			List: uis,
		}
		var roomParam *mdroom.RunCardRoomParam
		err := json.Unmarshal([]byte(mdr.GameParam), &roomParam)
		if err := json.Unmarshal([]byte(mdr.GameParam), &roomParam); err != nil {
			log.Err("create game unmarshal game param room_id:%d,game_param:%s ,err:%+v", mdr.RoomID,
				mdr.GameParam, err)
			continue
		}
		game := &mdgame.Runcard{
			RoomID:      mdr.RoomID,
			Status:      enumgame.GameStatusInit,
			Index:       1,
			PassWord:    mdr.Password,
			RoomParam:   roomParam.Option,
			GameResult:  gr,
			SubmitIndex: 0,
			OpDateAt:    &t,
		}
		mdr.Status = enumroom.RoomStatusStarted

		err = initUserCard(game, goLua)
		if err != nil {
			cachelog.SetErrLog(enumgame.ServiceCode, err.Error())
			continue
		}
		startCardType := "4_3"
		switch game.RoomParam[0][0] {
		case "1":
			startCardType = "4_3"
			break
		case "2":
			startCardType = "3_3"
			break
		case "3":
			startCardType = "2_3"
			break
		case "4":
			startCardType = "1_3"
			break
		default:
			startCardType = "4_3"
		}

		for _, gr := range game.GameResult.List {
			for _, card := range gr.CardList {
				if card == startCardType {
					game.OpUserID = gr.UserID
					game.FirstCardType = startCardType
				}
			}
		}

		if game.OpUserID == 0 {
			userIndex := tools.GenerateRangeNum(0, len(game.GameResult.List))
			game.OpUserID = game.GameResult.List[userIndex].UserID
		}

		index := game.GetUserInfoWithIndex(game.OpUserID)
		for i, ui := range game.GameResult.List {
			if i < index {
				ui.Index = int32(index+i) + 1
			} else {
				ui.Index = int32(i-index) + 1
			}
			//fmt.Printf("AAAACreateGame:%d|%d\n", ui.UserID, ui.Index)
		}

		f := func(tx *gorm.DB) error {
			if mdr.RoundNow == 1 {
				for _, user := range mdr.Users {
					pr := &mdroom.PlayerRoom{
						UserID:    user.UserID,
						RoomID:    mdr.RoomID,
						GameType:  mdr.GameType,
						PlayTimes: 0,
					}
					dbroom.CreatePlayerRoom(tx, pr)
				}
			}
			_, err := dbroom.UpdateRoom(tx, mdr)
			if err != nil {
				log.Err("run card room update set session failed, roomid:%d,err:%+v", mdr.RoomID, err)
				return err
			}

			err = dbgame.CreateGame(tx, game)
			if err != nil {
				log.Err("run card create set session failed,roomid:%d, err:%+v", mdr.RoomID, err)
				return err
			}

			err = cacheroom.UpdateRoom(mdr)
			if err != nil {
				log.Err("run update set session failed,roomid:%d,err: %+v", mdr.RoomID, err)
				return err
			}

			err = cachegame.SetGame(game)
			if err != nil {
				log.Err("run card create set redis failed,%v | %+v", mdr, err)
				return err
			}
			return nil
		}
		err = db.Transaction(f)
		if err != nil {
			cachelog.SetErrLog(enumgame.ServiceCode, err.Error())
			log.Err("run card create failed,%v | %+v", game, err)
			continue
		}
		newGames = append(newGames, game)
	}
	return newGames
}

//游戏流程更新逻辑
func UpdateGame(goLua *lua.LState) []*mdgame.Runcard {
	games, err := cachegame.GetAllGameByStatus(enumgame.GameStatusSubmitCardOver)
	if err != nil {
		log.Err("run card get all game by status failed, %v", err)
		return nil
	}
	if len(games) == 0 {
		return nil
	}
	//游戏结算结果集合
	var outGames []*mdgame.Runcard
	var specialCardUids []int32
	for _, game := range games {
		//if game.Status == enumgame.GameStatusInit {
		//	sub := time.Now().Sub(*game.OpDateAt)
		//	if sub.Seconds() > enumgame.SubmitCardTime {
		//		err := autoSubmitCard(game, goLua)
		//		if err != nil {
		//			log.Err("run card game auto submit card failed, roomid:%d,pwd:%s,err:%v", game.RoomID,
		//				game.PassWord, err)
		//			continue
		//		}
		//		outGames = append(outGames, game)
		//	}
		//} else
		if game.Status == enumgame.GameStatusSubmitCardOver {
			mdr, err := cacheroom.GetRoom(game.PassWord)
			if err != nil {
				log.Err("run card game status all submit card room get session failed, roomid:%d,pwd:%s,err:%v",
					game.RoomID, game.PassWord, err)
				continue
			}
			if mdr == nil {
				log.Err("run card room status all submit get session nil, %v|%d", game.PassWord, game.RoomID)
				continue
			}
			game.GameResult.IsSpring = enumgame.IsNom
			gameCardNum := tools.StringParseInt(game.RoomParam[0][0])
			noSubmitNum := 0
			var winerID int32 = 0
			for _, ur := range game.GameResult.List {
				ur.Status = enumgame.UserStatusDone
				cardNum := int32(len(ur.CardList))
				score := -cardNum
				score += ur.BoomScore
				if cardNum == gameCardNum {
					noSubmitNum ++
					ur.IsSpring = enumgame.IsClose
				} else if cardNum == 0 {
					winerID = ur.UserID
				}
			}
			if noSubmitNum == len(game.GameResult.List)-1 {
				game.GameResult.IsSpring = enumgame.IsSpring
				game.GetUserInfo(winerID).IsSpring = enumgame.IsSpring
			}

			//game.GameResult = results
			for _, result := range game.GameResult.List {
				for _, userResult := range mdr.UserResults {
					if userResult.UserID == result.UserID {
						ts := result.Score
						userResult.Score += ts
						if ts > 0 {
							userResult.Win += 1
						} else if ts == 0 {
							userResult.Tie += 1
						} else if ts < 0 {
							userResult.Lost += 1
						}
						userResult.RoundScore = ts
					}
				}
			}
			game.Status = enumgame.GameStatusDone
			mdr.Status = enumroom.RoomStatusReInit

			if mdr.RoomType == enumroom.RoomTypeClub && mdr.SubRoomType == enumroom.SubTypeClubMatch {
				err := room.GetRoomClubCoin(mdr)
				if err != nil {
					log.Err("room club member game balance failed,rid:%d,uid:%d, err:%v", mdr.RoomID, err)
					continue
				}
				for _, ur := range mdr.UserResults {
					ru := mdr.GetRoomUser(ur.UserID)
					if ru.UserRole != enumroom.UserRolePlayerBro {
						continue
					}
					for _, ugr := range game.GameResult.List {
						if ugr.UserID == ugr.UserID {
							ugr.ClubCoinScore = ur.RoundClubCoinScore
							break
						}
					}
				}
			}

			f := func(tx *gorm.DB) error {
				game, err = dbgame.UpdateGame(tx, game)
				if err != nil {
					log.Err("run card update db failed, %v|%v", game, err)
					return err
				}
				mdr, err = dbroom.UpdateRoom(tx, mdr)
				if err != nil {
					log.Err("run card update room db failed, %v|%v", game, err)
					return err
				}
				for _, uid := range specialCardUids {
					plsgr := &mdroom.PlayerSpecialGameRecord{
						GameID:     game.GameID,
						RoomID:     game.RoomID,
						GameType:   mdr.GameType,
						RoomType:   mdr.RoomType,
						Password:   mdr.Password,
						UserID:     uid,
						GameResult: game.GameResultStr,
					}
					err = dbroom.CreateSpecialGame(tx, plsgr)
					if err != nil {
						return err
					}
				}
				return nil
			}
			err = db.Transaction(f)
			if err != nil {
				log.Err("run card update failed, %v", err)
				continue
			}
			err = cachegame.DeleteGame(game)
			if err != nil {
				log.Err("run card room del session failed, roomid:%d,pwd:%s,err:%v", game.RoomID, game.PassWord, err)
				continue
			}

			err = cacheroom.UpdateRoom(mdr)
			if err != nil {
				log.Err("run card room update room redis failed,%v | %v",
					mdr, err)
				continue
			}
			outGames = append(outGames, game)
		}
	}
	return outGames
}

func initUserCard(game *mdgame.Runcard, goLua *lua.LState) error {
	//rp ,_ := json.Marshal(game.RoomParam)
	str := fmt.Sprintf("return G_GetCards(%s,%d)", game.RoomParam[0][0], len(game.GameResult.List))
	if err := goLua.DoString(str); err != nil {
		log.Err("run card G_GetCards err %v", err)
		return errgame.ErrGoLua
	}
	getCards := goLua.Get(-1)
	goLua.Pop(1)
	var cardArray [][]string
	if err := json.Unmarshal([]byte(getCards.String()), &cardArray); err != nil {
		return err
	}

	for i, gr := range game.GameResult.List {
		gr.CardList = cardArray[i]
		gr.CardNum = int32(len(cardArray[i]))
		gr.Score = 0
	}
	return nil
}

func autoSubmitCard(game *mdgame.Runcard, goLua *lua.LState) error {
	_, err := cacheroom.GetRoom(game.PassWord)
	if err != nil {
		return err
	}
	//game.MarshalGameResult()
	//if err := goLua.DoString(fmt.Sprintf("return G_CheckSubmitCard('%s','%s')",
	//	mdr.RoomParam, game.GameResultStr)); err != nil {
	//	log.Err("run card G_CalculateRes err %+v|%v\n", game.GameResultStr, err)
	//}
	//getResult := goLua.Get(-1)
	//goLua.Pop(1)
	//var results *mdgame.GameResult
	//if err := json.Unmarshal([]byte(getResult.String()), &results); err != nil {
	//	log.Err("run card lua str do struct %v", err)
	//}
	//game.GameResult = results
	//
	//err = cachegame.UpdateGame(game)
	//if err != nil {
	//	log.Err("run card set session failed, %v", err)
	//	return err
	//}
	return nil
}

func SubmitCard(uid int32, pwd string, cardList []string, goLua *lua.LState) (*mdgame.Runcard, []int32, error) {
	mdr, err := cacheroom.GetRoom(pwd)
	if err != nil {
		return nil, nil, err
	}

	if mdr.Status > enumroom.RoomStatusStarted {
		if mdr.Giveup == enumroom.WaitGiveUp {
			return nil, nil, errroom.ErrInGiveUp
		}
		return nil, nil, errroom.ErrGameIsDone
	}
	game, err := cachegame.GetGame(mdr.RoomID)
	if game == nil {
		return nil, nil, errgame.ErrGameNoFind
	}

	if game.Status != enumgame.GameStatusInit {
		return nil, nil, errgame.ErrSubmitCardDone
	}

	if game.OpUserID != uid {
		return nil, nil, errgame.ErrNotYourTurn
	}
	userResult := game.GetUserInfo(uid)

	if userResult == nil {
		return nil, nil, errgame.ErrUserNotInGame
	}

	if userResult.UserID != game.OpUserID {
		return nil, nil, errgame.ErrAlreadySubmitCard
	}

	if cacheroom.GetRoomTestConfigKey("RunCheckHasCards") != "0" {
		checkCards := game.GetUserInfo(uid).CardList
		checkHasCard := checkHasCards(cardList, checkCards)
		if !checkHasCard {
			return nil, nil, errgame.ErrCardNotExist
		}
	}
	err = checkSubmitCardOption(cardList,userResult,game)
	if err != nil{
		return nil,nil,err
	}
	sort.Strings(userResult.CardList)
	lastSubmitCardStr := ""
	if game.LastSubmitCard != nil {
		data, _ := json.Marshal(game.LastSubmitCard.CardList)
		lastSubmitCardStr = string(data)
	}

	cardListLen := int32(len(cardList))
	if cardListLen == 0 && game.LastSubmitCard == nil {
		return nil, nil, errgame.ErrSubmitCardNil
	}

	var submitType int32 = 1
	if cardListLen == userResult.CardNum {
		submitType = 2
	}

	m := game.GetUserMapByIndex(uid)
	otherSubmitCardStr, _ := json.Marshal(&m)
	roomParaStr, _ := json.Marshal(game.RoomParam)
	hasLastSubmit := 1
	if game.LastSubmitCard == nil {
		hasLastSubmit = 2
	}
	submitCardStr := ""
	if len(cardList) > 0 {
		data, _ := json.Marshal(cardList)
		submitCardStr = string(data)
	} else {
		data, _ := json.Marshal(userResult.CardList)
		submitCardStr = string(data)
	}
	if err := goLua.DoString(fmt.Sprintf("return SubmitCards('%s','%s','%s','%s','%s','%s')", submitType,
		hasLastSubmit, lastSubmitCardStr, string(submitCardStr), string(otherSubmitCardStr), string(roomParaStr)));
		err != nil {
		log.Err("run card G_CompareCards err %+v|%v\n", game.GameResultStr, err)
		return nil, nil, errgame.ErrSubmitCardFail
	}

	getResult := goLua.Get(-1)
	goLua.Pop(1)
	var result [][]int32
	if err := json.Unmarshal([]byte(getResult.String()), &result); err != nil {
		log.Err("run card lua str do struct %v", err)
	}
	compareResult := result[0][0]

	if compareResult == 2 {
		return nil, nil, errgame.ErrSubmitCardValueTooSmall
	} else if compareResult == -1 {
		return nil, nil, errgame.ErrSubmitCardType
	}

	var newCardList []string
	for _, card := range userResult.CardList {
		inList := false
		for _, submitCard := range cardList {
			if submitCard == card {
				inList = true
				break
			}
		}
		if !inList {
			newCardList = append(newCardList, card)
		}
	}
	cardType := result[1][0]

	nextUserID := result[2][0] //tools.StringParseInt()
	if nextUserID == 0 {
		nextUserID = uid
	}

	var passUserIds []int32
	for _, puid := range result[3] {
		passUserIds = append(passUserIds, puid) //tools.StringParseInt()
	}

	userResult.CardList = newCardList
	userResult.HasSubmitCard = 1
	userResult.CardNum = int32(len(userResult.CardList))

	submitCard := &mdgame.SubmitCard{}
	submitCard.CardList = cardList
	submitCard.UserID = uid
	submitCard.CardType = tools.IntToString(cardType)
	submitCard.Result = tools.IntToString(compareResult)
	submitCard.NextUserID = nextUserID
	game.LastSubmitCard = submitCard

	for _, boomType := range enumgame.RunCardBoomType {
		if boomType == submitCard.CardType {
			userResult.BoomNum ++
			userResult.BoomScore += enumgame.BoomType
			break
		}
	}

	//下一个玩家还是本来玩家 清空上家牌
	if game.OpUserID == nextUserID {
		for _, ur := range game.GameResult.List {
			ur.LastCards = nil
		}
	}
	game.OpUserID = nextUserID
	if len(userResult.CardList) == 0 {
		game.Status = enumgame.GameStatusSubmitCardOver
	}

	t := time.Now()
	game.OpDateAt = &t
	game.SubmitIndex ++
	//err = cachegame.UpdateGame(game)
	//if err != nil {
	//	log.Err("run card set session failed, %v", err)
	//	return nil, nil, err
	//}
	return game, passUserIds, nil
}

func checkSubmitCardOption(cardList []string, userInfo *mdgame.UserInfo,game *mdgame.Runcard) error {
	startCardType := "4_3"
	switch game.RoomParam[0][0] {
	case "1":
		startCardType = "4_3"
		break
	case "2":
		startCardType = "3_3"
		break
	case "3":
		startCardType = "2_3"
		break
	case "4":
		startCardType = "1_3"
		break
	default:
		startCardType = "4_3"
	}
	hasStartCardType := false
	for _,card := range userInfo.CardList{
		if card == startCardType{
			hasStartCardType = true
			break
		}
	}
	if !hasStartCardType{
		return nil
	}
	hasStartCardType = false
	for _,card := range cardList{
		if card == startCardType{
			hasStartCardType = true
			break
		}
	}
	if !hasStartCardType{
		return errgame.ErrStartCardSubmit
	}
	return nil
}

func GameResultList(rid int32) (*pbgame.GameResultListReply, error) {
	var list []*pbgame.GameResult
	games, err := dbgame.GetRunCardByRoomID(db.DB(), rid)
	if err != nil {
		return nil, err
	}
	for _, game := range games {
		list = append(list, game.ToProto())
	}
	out := &pbgame.GameResultListReply{
		List: list,
	}
	return out, nil
}

func GameRecovery(rid int32) (*mdgame.Runcard, error) {
	game, err := cachegame.GetGame(rid)
	if err != nil {
		return nil, err
	}
	if game == nil {
		game, err = dbgame.GetLastRunCardByRoomID(db.DB(), rid)
		if err != nil {
			return nil, err
		}
	}
	if game == nil {
		return nil, nil
	}
	return game, nil
}

func GameExist(uid int32, rid int32) (*pbgame.RecoveryReply, error) {
	out := &pbgame.RecoveryReply{}
	_, roomRecovery, err := room.CheckRoomExist(uid, rid)
	if err != nil {
		return nil, err
	}
	out.RoomExist = roomRecovery.ToProto()
	out.RoomExist.Room.CreateOrEnter = enumroom.EnterRoom
	if len(out.RoomExist.Room.UserList) > 0 {
		out.RoomExist.Room.OwnerID = out.RoomExist.Room.UserList[0].UserID
	}
	//客户端需要 游戏开始前 RoundNow都算上一局的（玩家挂起时游戏恢复看不到刚挂起那个人的牌）
	if out.RoomExist.Room.RoundNow > 1 && out.RoomExist.Room.Status < enumroom.RoomStatusStarted {
		out.RoomExist.Room.RoundNow -= 1
	}
	if roomRecovery.Status < enumroom.RecoveryGameStart && roomRecovery.Status != enumroom.RecoveryInitNoReady {
		return out, nil
	}
	game, err := GameRecovery(roomRecovery.Room.RoomID)
	if err != nil {
		return nil, err
	}
	if game == nil {
		return out, err
	}
	out.RunCardExist = game.ToProto()
	for _, gr := range out.RunCardExist.List {
		if gr.UserID != uid && game.Status < enumgame.GameStatusDone {
			gr.CardList = nil
		}
	}
	var time int32
	switch game.Status {
	case enumgame.GameStatusInit:
		time = enumgame.SubmitCardTime
		break
	}
	out.CountDown = &pbgame.CountDown{
		ServerTime: game.OpDateAt.Unix(),
		Count:      time,
	}
	return out, nil
}

func CleanGame() error {
	var gids []int32
	rids, err := cacheroom.GetAllDeleteRoomKey(enumroom.RunCardGameType)
	if err != nil {
		log.Err("get run card clean room err:%v", err)
		return err
	}
	for _, rid := range rids {
		game, err := cachegame.GetGame(rid)
		if err != nil {
			log.Err("get run card give up room err:%d|%v", rid, err)
			continue
		}
		if game != nil {
			log.Debug("clean run card game:%d|%d\n", game.GameID, game.RoomID)
			gids = append(gids, game.GameID)
			err = cachegame.DeleteGame(game)
			if err != nil {
				log.Err(" delete run card set session failed, %v",
					err)
				continue
			}
			err = cacheroom.CleanDeleteRoom(enumgame.GameID, game.RoomID)
			if err != nil {
				log.Err(" delete run card delete room session failed,roomid:%d,err: %v", game.RoomID,
					err)
				continue
			}
		} else {
			err = cacheroom.CleanDeleteRoom(enumgame.GameID, rid)
			if err != nil {
				log.Err(" delete null game run card delete room session failed,roomid:%d,err: %v", rid,
					err)
				continue
			}
		}
	}
	if len(gids) > 0 {
		f := func(tx *gorm.DB) error {
			err = dbgame.GiveUpGameUpdate(tx, gids)
			if err != nil {
				return err
			}
			return nil
		}
		go db.Transaction(f)
	}

	return nil
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
