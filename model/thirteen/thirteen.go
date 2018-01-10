package thirteen

import (
	"encoding/json"
	"fmt"
	cacher "playcards/model/room/cache"
	dbr "playcards/model/room/db"
	enumr "playcards/model/room/enum"
	"playcards/model/room/errors"
	mdr "playcards/model/room/mod"
	cachet "playcards/model/thirteen/cache"
	dbt "playcards/model/thirteen/db"
	enumt "playcards/model/thirteen/enum"
	errorst "playcards/model/thirteen/errors"
	mdt "playcards/model/thirteen/mod"
	pbt "playcards/proto/thirteen"
	enumbill "playcards/model/bill/enum"
	mdbill "playcards/model/bill/mod"
	"playcards/model/room"
	//"playcards/utils/env"
	"playcards/model/bill"
	"playcards/utils/db"
	"playcards/utils/log"
	"strconv"
	"github.com/jinzhu/gorm"
	"github.com/yuin/gopher-lua"
)

//var GoLua *lua.LState
//
//func InitGoLua(gl *lua.LState) {
//	GoLua = gl
//}

func CreateThirteen(goLua *lua.LState) []*mdt.Thirteen {
	mdrooms := cacher.GetAllRoomByGameTypeAndStatus(enumr.ThirteenGameType, enumr.RoomStatusAllReady)
	if mdrooms == nil && len(mdrooms) == 0 {
		return nil
	}

	if len(mdrooms) == 0 {
		return nil
	}
	var newGames []*mdt.Thirteen

	for _, mdroom := range mdrooms {
		var userResults []*mdr.GameUserResult
		var groupCards []*mdt.GroupCard
		var bankerID int32
		if err := goLua.DoString("return G_Reset()"); err != nil {
			log.Err("thirteen G_Reset %+v", err)
			continue
		}
		for _, user := range mdroom.Users {
			if mdroom.RoundNow == 1 {
				userResult := &mdr.GameUserResult{
					UserID:   user.UserID,
					//Nickname: user.Nickname,
					Role:     user.Role,
					Win:      0,
					Lost:     0,
					Tie:      0,
					Score:    0,
				}
				userResults = append(userResults, userResult)
			}
			if err := goLua.DoString("return G_GetCards()"); err != nil {
				log.Err("thirteen G_GetCards %+v", err)
			}
			getCards := goLua.Get(-1)
			goLua.Pop(1)
			var cardList []string
			if cardsMap, ok := getCards.(*lua.LTable); ok {
				cardsMap.ForEach(func(key lua.LValue, value lua.LValue) {
					if cards, ok := value.(*lua.LTable); ok {
						var cardType string
						var cardValue string
						cards.ForEach(func(k lua.LValue, v lua.LValue) {
							if k.String() == "_type" {
								cardType = v.String()
							} else {
								cardValue = v.String()
							}
						})
						cardList = append(cardList, cardType+"_"+cardValue)
					} else {
						log.Err("thirteen cardsMap value err %+v", value)
					}
				})
				groupCard := &mdt.GroupCard{
					UserID:     user.UserID,
					CardList:   cardList,
					RoomStatus: enumr.RoomStatusStarted,
				}
				groupCards = append(groupCards, groupCard)
			} else {
				log.Err("thirteen cardsMap err %+v", cardsMap)
			}
			if user.Role == enumr.UserRoleMaster {
				bankerID = user.UserID
			}
		}

		if mdroom.RoundNow == 1 {
			mdroom.UserResults = userResults
		}
		thirteen := &mdt.Thirteen{
			RoomID:      mdroom.RoomID,
			BankerID:    bankerID,
			Status:      enumt.GameStatusInit,
			Index:       mdroom.RoundNow,
			PassWord:    mdroom.Password,
			SubmitCards: []*mdt.SubmitCard{},
			//GameLua: l,
			Cards: groupCards,
			Ids:   mdroom.Ids,
		}
		mdroom.Status = enumr.RoomStatusStarted
		f := func(tx *gorm.DB) error {
			if mdroom.RoundNow == 1 {
				if mdroom.RoomType != enumr.RoomTypeClub && mdroom.Cost != 0 {
					err := bill.GainGameBalance(mdroom.PayerID, mdroom.RoomID, enumbill.JournalTypeThirteen,
						enumbill.JournalTypeThirteenUnFreeze, &mdbill.Balance{Amount: -mdroom.Cost, CoinType: mdroom.CostType})
					if err != nil {
						log.Err("thirteen create user balance failed,%v | %+v", mdroom.PayerID, err)
						return err
					}
				}

				for _, user := range mdroom.Users {
					pr := &mdr.PlayerRoom{
						UserID:    user.UserID,
						RoomID:    mdroom.RoomID,
						GameType:  mdroom.GameType,
						PlayTimes: 0,
					}
					dbr.CreatePlayerRoom(tx, pr)
				}
			}

			_, err := dbr.UpdateRoom(tx, mdroom)
			if err != nil {
				log.Err("thirten room update set session failed, roomid:%d,err:%+v", mdroom.RoomID, err)
				return err
			}

			err = dbt.CreateThirteen(tx, thirteen)
			if err != nil {
				log.Err("thirteen create set session failed,roomid:%d, err:%+v", mdroom.RoomID, err)
				return err
			}

			err = cacher.UpdateRoom(mdroom)
			if err != nil {
				log.Err("room update set session failed,roomid:%d,err: %+v", mdroom.RoomID, err)
				return err
			}

			err = cachet.SetGame(thirteen)
			if err != nil {
				log.Err("thirteen create set redis failed,%v | %+v", mdroom, err)
				return err
			}

			//for _, user := range room.Users {
			//	err = cachet.SetGameUser(room.RoomID, user.UserID)
			//	if err != nil {
			//		log.Err("thirteen create set game user redis failed,%v | %+v", room, err)
			//		return err
			//	}
			//}

			return nil
		}
		//go db.Transaction(f)

		err := db.Transaction(f)
		if err != nil {
			log.Err("thirteen create failed,%v | %+v", thirteen, err)
			continue
		}

		//err = cacher.SetRoom(room)
		//if err != nil {
		//	log.Err("room create set redis failed,%v | %+v", room, err)
		//	continue
		//}
		newGames = append(newGames, thirteen)
	}
	return newGames
}

//游戏结算逻辑
func UpdateGame(goLua *lua.LState) []*mdt.Thirteen { //[]*mdt.GameResultList

	thirteens := cachet.GetAllThirteenByStatus(enumt.GameStatusStarted)
	if len(thirteens) == 0 {
		return nil
	}
	//游戏结算结果集合
	var thirteenList []*mdt.Thirteen
	//var resultListArray []*mdt.GameResultList

	for _, thirteen := range thirteens {
		resultList := &mdt.GameResultList{}
		//resultList.RoomID = thirteen.RoomID
		thirteen.Status = enumt.GameStatusDone
		var bankerScore int32
		//var resultArray []*mdt.GameResult

		//获取游戏所属房间缓存 更新房间信息
		mdroom, err := cacher.GetRoom(thirteen.PassWord)
		if err != nil {
			log.Err("room get session failed, %v", err)
			err = cachet.DeleteGame(thirteen)
			if err != nil {
				log.Err("thirteen room not exist delete  set session failed, %v",
					err)
			}
			log.Err("thirteen room not exist delete  game, %v",
				thirteen)
			continue
			continue
		}
		if mdroom == nil {
			log.Err("room get session failed, %v", err)
			continue
		}

		var results []*mdt.ThirteenResult

		thirteen.MarshalUserSubmitCards()
		//fmt.Printf("GameParam:%v\n%v\n",thirteen.UserSubmitCards,room.GameParam)
		if err := goLua.DoString(fmt.Sprintf("return G_GetResult('%s','%s')",
			thirteen.UserSubmitCards, mdroom.GameParam)); err != nil {
			log.Err("thirteen G_GetResult submit card:%s, game param:%s,%v", thirteen.UserSubmitCards, mdroom.GameParam, err)
			continue
		}

		luaResult := goLua.Get(-1)
		goLua.Pop(1)
		if err := json.Unmarshal([]byte(luaResult.String()), &results); err != nil {
			log.Err("thirteen lua str do struct %v", err)
			continue
		}

		resultList.Result = results

		var resultArray []*mdr.GameUserResult
		for _, result := range resultList.Result {
			for _, userResult := range mdroom.UserResults {
				m := initThirteenGameTypeMap()
				if userResult.UserID == result.UserID {
					if len(userResult.GameCardCount) > 0 {
						if err := json.Unmarshal([]byte(userResult.GameCardCount), &m); err != nil {
							log.Err("thirteen game card count lua str do struct %v", err)
						}
					}
					ts, _ := strconv.ParseInt(result.Settle.TotalScore, 10, 32)
					userResult.Score += int32(ts)

					if ts > 0 {
						userResult.Win += 1
					} else if ts == 0 {
						userResult.Tie += 1
					} else if ts < 0 {
						userResult.Lost += 1
					}
					if _, ok := m[result.Result.Head.GroupType]; ok {
						m[result.Result.Head.GroupType]++
					}
					if _, ok := m[result.Result.Middle.GroupType]; ok {
						m[result.Result.Middle.GroupType]++
					}
					if _, ok := m[result.Result.Tail.GroupType]; ok {
						m[result.Result.Tail.GroupType]++
					}
					m["Shoot"] += int32(len(result.Result.Shoot))
					if len(resultList.Result) > 2 &&
						len(result.Result.Shoot) >= (len(resultList.Result)-1) {
						m["AllShoot"]++
					}
					r, _ := json.Marshal(m)
					userResult.GameCardCount = string(r)
					resultArray = append(resultArray, userResult)
					if userResult.UserID == thirteen.BankerID {
						bankerScore = int32(ts)
					}
				}
			}
		}

		//resultListArray = append(resultListArray, resultList)
		thirteen.Result = resultList
		mdroom.Status = enumr.RoomStatusReInit

		var roomparam *mdr.ThirteenRoomParam
		if err := json.Unmarshal([]byte(mdroom.GameParam), &roomparam); err != nil {
			log.Err("thirteen clean unmarshal room param failed, %v", err)
			continue
		}

		if roomparam.BankerAddScore > 0 && roomparam.Times != enumt.TimesDefault {
			for i := 0; i < len(mdroom.Users); i++ {
				//room.Users[i].Ready = enumr.UserUnready
				//十三张一局结束后 轮庄
				if mdroom.Users[i].Role == enumr.UserRoleMaster {
					if bankerScore <= 0 {
						if i == len(mdroom.Users)-1 {
							mdroom.Users[0].Role = enumr.UserRoleMaster
						} else {
							mdroom.Users[i+1].Role = enumr.UserRoleMaster
						}
						mdroom.Users[i].Role = enumr.UserRoleSlave
					}
					break
				}
			}
		}

		f := func(tx *gorm.DB) error {
			//fmt.Printf("UpdateThirteen:%+v", thirteen)
			thirteen, err = dbt.UpdateThirteen(tx, thirteen)
			if err != nil {
				log.Err("thirteen update failed, %v\n", err)
				return err
			}
			r, err := dbr.UpdateRoom(tx, mdroom)
			if err != nil {
				return err
			}
			mdroom = r
			return nil
		}
		//go db.Transaction(f)
		err = db.Transaction(f)
		if err != nil {
			//print(err)
			log.Err("thirteen update failed, %v", err)
			return nil
		}
		err = cachet.DeleteGame(thirteen)
		if err != nil {
			log.Err("thirteen set session failed, %v", err)
			return nil
		}
		err = cacher.UpdateRoom(mdroom)
		if err != nil {
			log.Err("room update room redis failed,%v | %v", mdroom, err)
			return nil
		}
		thirteenList = append(thirteenList, thirteen)
	}
	return thirteenList
}

func initThirteenGameTypeMap() map[string]int32 {
	m := make(map[string]int32)
	for _, value := range enumt.GroupTypeName {
		m[value] = 0
	}
	return m
}

func SubmitCard(uid int32, submitCard *mdt.SubmitCard, mdroom *mdr.Room) ([]int32, error) {
	if mdroom.Status > enumr.RoomStatusStarted {
		return nil, errors.ErrGameIsDone
	}
	if mdroom.Giveup == enumr.WaitGiveUp {
		return nil, errors.ErrInGiveUp
	}

	thirteen, err := cachet.GetGame(mdroom.RoomID)
	if err != nil {
		return nil, err
	}
	if thirteen == nil {
		return nil, errorst.ErrGameNotExist
	}

	for _, uc := range thirteen.SubmitCards {
		if uc.UserID == uid {
			return nil, errorst.ErrAlreadySubmitCard
		}
	}
	if cacher.GetRoomTestConfigKey("ThirteenCheckHasCards") != "0" {
		var checkCards []string
		for _, cardGroup := range thirteen.Cards {
			if cardGroup.UserID == uid {
				checkCards = cardGroup.CardList
			}
		}

		checkHasCard := checkHasCards(submitCard, checkCards)
		if !checkHasCard {
			return nil, errorst.ErrCardNotExist
		}
	}

	for _, user := range mdroom.Users {
		if user.UserID == uid {
			submitCard.Role = user.Role
		}
	}

	submitCard.UserID = uid
	thirteen.SubmitCards = append(thirteen.SubmitCards, submitCard)
	if thirteen.Status > enumt.GameStatusInit {
		return nil, errors.ErrGameIsDone
	}

	if mdroom.MaxNumber == int32(len(thirteen.SubmitCards)) {
		thirteen.Status = enumt.GameStatusStarted
	}

	//f := func(tx *gorm.DB) error {
	//	thirteen, err = dbt.UpdateThirteen(tx, thirteen)
	//	if err != nil {
	//		return err
	//	}
	//	return nil
	//}
	//err = db.Transaction(f)
	//if err != nil {
	//	return 0, err
	//}
	//err = cachet.UpdateGame(thirteen)
	//if err != nil {
	//	log.Err("thirteen update session failed, %v", err)
	//	return 0, err
	//}

	err = cachet.UpdateGame(thirteen)
	if err != nil {
		log.Err("thirteen set session failed, %v", err)
		return nil, err
	}
	return thirteen.Ids, nil //

}

//func GetThirteenByStatusAndGameType() ([]*mdt.Thirteen, error) {
//	var (
//		thirteens []*mdt.Thirteen
//	)
//	list, err := dbt.GetThirteensByStatus(db.DB(),
//		enumt.GameStatusStarted)
//	if err != nil {
//		return nil, err
//	}
//	thirteens = list
//	return thirteens, nil
//}

func GameResultList(rid int32) (*pbt.GameResultListReply, error) {
	var list []*pbt.GameResultList
	thirteens, err := dbt.GetThirteenByRoomID(db.DB(), rid)
	if err != nil {
		return nil, err
	}
	for _, thirteen := range thirteens {
		result := thirteen.Result
		list = append(list, result.ToProto())
	}
	out := &pbt.GameResultListReply{
		List: list,
	}
	return out, nil
}

func CleanGame() error {
	var gids []int32
	//rids, err := dbr.GetGiveUpRoomIDByGameType(db.DB(), enumt.GameID)
	//if err != nil {
	//	log.Err("get thirteen give up room err:%v", err)
	//}

	rids, err := cacher.GetAllDeleteRoomKey(enumr.ThirteenGameType)
	if err != nil {
		log.Err("get thirteen clean room err:%v", err)
		return err
	}
	for _, rid := range rids {
		game, err := cachet.GetGame(rid)
		if err != nil {
			log.Err("get thirteen give up room err:%d|%v", rid, err)
			continue
		}
		if game != nil {
			gids = append(gids, game.GameID)
			err = cachet.DeleteGame(game)
			//log.Debug("clean thirteen game:%d|%d|%+v\n",game.GameID,game.RoomID,game.Ids)
			if err != nil {
				log.Err(" delete thirteen set session failed, %v", err)
				continue
			}
			err = cacher.CleanDeleteRoom(enumt.GameID, game.RoomID)
			if err != nil {
				log.Err(" delete thirteen delete room session failed,roomid:%d,err: %v", game.RoomID,
					err)
				continue
			}
		} else {
			err = cacher.CleanDeleteRoom(enumt.GameID, rid)
			if err != nil {
				log.Err(" delete null game thirteen delete room session failed,roomid:%d,err: %v", rid,
					err)
				continue
			}
		}
	}
	if len(gids) > 0 {
		f := func(tx *gorm.DB) error {
			err = dbt.GiveUpGameUpdate(tx, gids)
			if err != nil {
				return err
			}
			return nil
		}
		go db.Transaction(f)
		//err = db.Transaction(f)
		//if err != nil {
		//	return err
		//}
	}

	return nil
}

func ThirteenRecovery(rid int32, uid int32) (*mdt.ThirteenRecovery, error) {
	thirteen, err := cachet.GetGame(rid)
	if err != nil {
		return nil, err
	}
	if thirteen == nil {
		thirteen, err = dbt.GetLastThirteenByRoomID(db.DB(), rid)
		if err != nil {
			return nil, err
		}
	}
	if thirteen == nil {
		return nil, errorst.ErrGameNotExist
	}
	recovery := &mdt.ThirteenRecovery{}
	var readyuser []int32
	gameStatus := enumt.RecoveryInitNoReady
	if thirteen.Status == enumt.GameStatusInit {
		for _, card := range thirteen.Cards {
			if card.UserID == uid {
				recovery.Cards = *card
			}
		}
		for _, card := range thirteen.SubmitCards {
			if card.UserID == uid {
				gameStatus = enumt.RecoveryInitReady
			}
			readyuser = append(readyuser, card.UserID)
		}
	} else {
		gameStatus = enumt.RecoveryGameStart
		recovery.GameResult = *thirteen.Result
	}
	recovery.Status = int32(gameStatus)
	recovery.ReadyUser = readyuser
	recovery.BankerID = thirteen.BankerID
	return recovery, nil
}

func ThirteenExist(uid int32, rid int32) (*pbt.ThirteenRecoveryReply, error) {

	out := &pbt.ThirteenRecoveryReply{}
	_, roomRecovery, err := room.CheckRoomExist(uid, rid)
	if err != nil {
		return nil, err
	}
	out.RoomExist = roomRecovery.ToProto()
	out.RoomExist.Room.CreateOrEnter = enumr.EnterRoom
	out.RoomExist.Room.OwnerID = out.RoomExist.Room.UserList[0].UserID
	if roomRecovery.Status < enumr.RecoveryGameStart && roomRecovery.Status != enumr.RecoveryInitNoReady {
		return out, nil
	}
	thirteen, err := cachet.GetGame(roomRecovery.Room.RoomID)
	if err != nil {
		return nil, err
	}
	if thirteen == nil {
		thirteen, err = dbt.GetLastThirteenByRoomID(db.DB(), roomRecovery.Room.RoomID)
		if err != nil {
			return nil, err
		}
	}
	if thirteen == nil {
		return nil, errorst.ErrGameNotExist
	}
	recovery := &mdt.ThirteenRecovery{}
	var readyuser []int32
	gameStatus := enumt.RecoveryInitNoReady
	//if thirteen.Status == enumt.GameStatusInit {
	//	for _, card := range thirteen.Cards {
	//		if card.UserID == uid {
	//			recovery.Cards = *card
	//		}
	//	}
	//	for _, card := range thirteen.SubmitCards {
	//		if card.UserID == uid {
	//			gameStatus = enumt.RecoveryInitReady
	//		}
	//		readyuser = append(readyuser, card.UserID)
	//	}
	//} else {
	//
	//	gameStatus = enumt.RecoveryGameStart
	//	recovery.GameResult = *thirteen.Result
	//}

	for _, card := range thirteen.Cards {
		if card.UserID == uid {
			recovery.Cards = *card
		}
	}
	for _, card := range thirteen.SubmitCards {
		if card.UserID == uid {
			gameStatus = enumt.RecoveryInitReady
		}
		readyuser = append(readyuser, card.UserID)
	}
	if thirteen.Status == enumt.GameStatusDone {
		gameStatus = enumt.RecoveryGameStart
		recovery.GameResult = *thirteen.Result
	}
	recovery.Status = int32(gameStatus)
	recovery.ReadyUser = readyuser
	recovery.BankerID = thirteen.BankerID
	//out := &pbt.ThirteenRecovery{
	//	RoomExist:     roomRecovery.ToProto(),
	//	ThirteenExist: recovery.ToProto(),
	//}
	out.ThirteenExist = recovery.ToProto()
	return out, nil
}

func checkHasCards(submitCards *mdt.SubmitCard, cardList []string) bool {
	var submitCardStr []string
	for _, sc := range submitCards.Head {
		submitCardStr = append(submitCardStr, sc)
	}
	for _, sc := range submitCards.Middle {
		submitCardStr = append(submitCardStr, sc)
	}
	for _, sc := range submitCards.Tail {
		submitCardStr = append(submitCardStr, sc)
	}

	ctemp := cardList

	for i := 0; i < len(submitCardStr); i++ {
		for j := 0; j < len(ctemp); j++ {
			if submitCardStr[i] == ctemp[j] {
				submitCardStr[i] = "ok"
				ctemp[j] = "pass"
				continue
			}
		}
	}

	for _, sc := range submitCardStr {
		if len(sc) > 2 {
			return false
		}
	}
	return true
}
