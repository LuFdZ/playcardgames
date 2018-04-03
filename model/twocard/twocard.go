package twocard

import (
	"encoding/json"
	"fmt"
	cacheroom "playcards/model/room/cache"
	dbroom "playcards/model/room/db"
	enumroom "playcards/model/room/enum"
	errroom"playcards/model/room/errors"
	mdroom "playcards/model/room/mod"
	cachegame "playcards/model/twocard/cache"
	dbgame "playcards/model/twocard/db"
	enumgame "playcards/model/twocard/enum"
	errgame "playcards/model/twocard/errors"
	mdgame "playcards/model/twocard/mod"
	pbtwo "playcards/proto/twocard"
	"playcards/model/room"
	"math/rand"
	"playcards/utils/db"
	"playcards/utils/log"
	"github.com/jinzhu/gorm"
	"github.com/yuin/gopher-lua"
	"time"
	"sort"
)

func CreateGame() []*mdgame.Twocard {
	rooms := cacheroom.GetAllRoomByGameTypeAndStatus(enumroom.TwoCardGameType, enumroom.RoomStatusAllReady)
	if rooms == nil && len(rooms) == 0 {
		return nil
	}
	if len(rooms) == 0 {
		return nil
	}
	var newGames []*mdgame.Twocard

	for _, mdr := range rooms {
		var (
			userResults []*mdroom.GameUserResult
			userInfo    []*mdgame.UserInfo
		)
		newGameResult := &mdgame.GameResult{}
		var bankerID int32
		bankerID = mdr.BankerList[0]
		for _, user := range mdr.Users {
			if mdr.RoundNow == 1 {
				userResult := &mdroom.GameUserResult{
					UserID: user.UserID,
					Role:   user.Role,
					Win:    0,
					Lost:   0,
					Tie:    0,
					Score:  0,
				}
				userResults = append(userResults, userResult)
			}
			//if user.Role == enumroom.UserRoleMaster {
			//	bankerID = user.UserID
			//}
			if user.UserID == bankerID {
				user.Role = enumroom.UserRoleMaster
			}else{
				user.Role = enumroom.UserRoleSlave
			}
			ui := &mdgame.UserInfo{
				UserID:     user.UserID,
				Status:     enumgame.UserStatusInit,
				Bet:        0,
				Role:       user.Role,
				TotalScore: 0,
			}
			userInfo = append(userInfo, ui)

		}
		newGameResult.List = userInfo
		if mdr.RoundNow == 1 {
			mdr.UserResults = userResults
		}
		now := gorm.NowFunc()

		var roomParam *mdroom.TwoCardRoomParam
		if err := json.Unmarshal([]byte(mdr.GameParam), &roomParam); err != nil {
			log.Err("create two card unmarshal room param failed, %v", err)
			continue
		}
		game := &mdgame.Twocard{
			RoomID:     mdr.RoomID,
			BankerID:   bankerID,
			Status:     enumgame.GameStatusInit,
			Index:      mdr.RoundNow,
			PassWord:   mdr.Password,
			GameResult: newGameResult,
			BetType:    roomParam.BetType,
			OpDateAt:   &now,
			Ids:        mdr.Ids,
		}
		mdr.Status = enumroom.RoomStatusStarted
		f := func(tx *gorm.DB) error {
			if mdr.RoundNow == 1 {
				//if mdroom.RoomType != enumroom.RoomTypeClub && mdroom.Cost != 0 {
				//	err := bill.GainGameBalance(mdroom.PayerID, mdroom.RoomID, enumbill.JournalTypeThirteen,
				//		enumbill.JournalTypeThirteenUnFreeze, &mdbill.Balance{Amount: -mdroom.Cost, CoinType: mdroom.CostType})
				//	if err != nil {
				//		log.Err("four card create user balance failed,%v | %+v", mdroom.PayerID, err)
				//		return err
				//	}
				//}
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
				log.Err("two card room update set session failed, roomid:%d,err:%+v", mdr.RoomID, err)
				return err
			}

			err = dbgame.CreateGame(tx, game)
			if err != nil {
				log.Err("two card create set session failed,roomid:%d, err:%+v", mdr.RoomID, err)
				return err
			}

			err = cacheroom.UpdateRoom(mdr)
			if err != nil {
				log.Err("two update set session failed,roomid:%d,err: %+v", mdr.RoomID, err)
				return err
			}

			err = cachegame.SetGame(game)
			if err != nil {
				log.Err("two card create set redis failed,%v | %+v", mdr, err)
				return err
			}
			return nil
		}
		//go db.Transaction(f)

		err := db.Transaction(f)
		if err != nil {
			log.Err("two card create failed,%v | %+v", game, err)
			continue
		}
		newGames = append(newGames, game)
	}
	return newGames
}

//游戏结算逻辑
func UpdateGame(goLua *lua.LState) []*mdgame.Twocard {
	games, err := cachegame.GetAllGameByStatus(enumgame.GameStatusDone)
	if err != nil {
		log.Err("two card get all game by status failed, %v", err)
		return nil
	}
	if len(games) == 0 {
		return nil
	}
	//游戏结算结果集合
	var outGames []*mdgame.Twocard
	var specialCardUids []int32
	for _, game := range games {
		if game.Status == enumgame.GameStatusInit {
			sub := time.Now().Sub(*game.OpDateAt)
			if sub.Seconds() > enumgame.SetBetTime || game.BetType == enumgame.BetTypeNo {
				autoSetBankerScore(game)
				game.Status = enumgame.GameStatusAllBet
				err := cachegame.UpdateGame(game)
				if err != nil {
					log.Err("two card game status init set session failed, %v", err)
					continue
				}
			}
		} else if game.Status == enumgame.GameStatusAllBet {
			randomUserDice(game)
			err := initUserCard(game, goLua)
			if err != nil {
				//print(err)
				log.Err("two card init user room get session failed, roomid:%d,pwd:%s,err:%v", game.RoomID, game.PassWord, err)
				continue
			}
			game.Status = enumgame.GameStatusOrdered
			err = cachegame.UpdateGame(game)
			if err != nil {
				log.Err("two card game status ordered set session failed, %v", err)
				continue
			}
		} else if game.Status == enumgame.GameStatusOrdered {
			game.Status = enumgame.GameStatusSubmitCard
			t := time.Now()
			game.OpDateAt = &t
			err = cachegame.UpdateGame(game)
			if err != nil {
				log.Err("two card user status submit card set session failed, %v", err)
				continue
			}
		} else if game.Status == enumgame.GameStatusSubmitCard {
			sub := time.Now().Sub(*game.OpDateAt)
			if sub.Seconds() > enumgame.SubmitCardTime{
				err := autoSubmitCard(game)
				if err != nil{
					log.Err("two card user all submit card set session failed, %v", err)
					continue
				}
			}

		} else if game.Status == enumgame.GameStatusAllSubmitCard {
			mdr, err := cacheroom.GetRoom(game.PassWord)
			if err != nil {
				log.Err("two card game status all submit card room get session failed, roomid:%d,pwd:%s,err:%v", game.RoomID, game.PassWord, err)
				continue
			}
			if mdr == nil {
				log.Err("two card room status all submit get session nil, %v|%d", game.PassWord, game.RoomID)
				continue
			}
			game.MarshalGameResult()
			if err := goLua.DoString(fmt.
			Sprintf("return G_CalculateRes('%s','%s')",
				game.GameResultStr, mdr.GameParam)); err != nil {
				log.Err("two card G_CalculateRes err %+v|\n%+v|\n%v\n",
					game.GameResultStr, mdr.GameParam, err)
				continue
			}

			luaResult := goLua.Get(-1)
			goLua.Pop(1)
			var results *mdgame.GameResult
			if err := json.Unmarshal([]byte(luaResult.String()),
				&results); err != nil {
				log.Err("two card lua str do struct %v", err)
			}
			game.GameResult = results
			for _, result := range game.GameResult.List {
				m := initTwoCardTypeMap()
				for _, userResult := range mdr.UserResults {
					if userResult.UserID == result.UserID {
						if len(userResult.GameCardCount) > 0 {
							if err := json.Unmarshal([]byte(userResult.GameCardCount), &m); err != nil {
								log.Err("two card lua str do struct %v", err)
							}
						}

						if _, ok := m[result.Cards.CardType]; ok {
							m[result.Cards.CardType]++
						}

						r, _ := json.Marshal(m)
						userResult.GameCardCount = string(r)
						ts := result.TotalScore
						userResult.Score += ts

						if ts > 0 {
							userResult.Win += 1
						} else if ts == 0 {
							userResult.Tie += 1
						} else if ts < 0 {
							userResult.Lost += 1
						}
						//特殊记录 至尊 天对
						if result.Cards.CardType == 100 || result.Cards.CardType == 110 {
							specialCardUids = append(specialCardUids, userResult.UserID)
						}
						userResult.RoundScore = ts
					}
				}
			}
			game.Status = enumgame.GameStatusDone
			mdr.Status = enumroom.RoomStatusReInit

			if mdr.RoomType == enumroom.RoomTypeClub && mdr.SubRoomType == enumroom.SubTypeClubMatch {
				err := room.GetRoomClubCoin(mdr)
				if err != nil{
					log.Err("room club member game balance failed,rid:%d,uid:%d, err:%v", mdr.RoomID, err)
					continue
				}
				for _,ur := range mdr.UserResults{
					for _,ugr := range game.GameResult.List{
						if ugr.UserID == ugr.UserID{
							ugr.ClubCoinScore = ur.RoundClubCoinScore
							break
						}
					}
				}
			}

			f := func(tx *gorm.DB) error {
				game, err = dbgame.UpdateGame(tx, game)
				if err != nil {
					log.Err("two card update db failed, %v|%v", game, err)
					return err
				}
				mdr, err = dbroom.UpdateRoom(tx, mdr)
				if err != nil {
					log.Err("two card update room db failed, %v|%v", game, err)
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
				log.Err("two card update failed, %v", err)
				continue
			}
			err = cachegame.DeleteGame(game)
			if err != nil {
				log.Err("two card room del session failed, roomid:%d,pwd:%s,err:%v", game.RoomID, game.PassWord, err)
				continue
			}

			err = cacheroom.UpdateRoom(mdr)
			if err != nil {
				log.Err("two card room update room redis failed,%v | %v",
					mdr, err)
				continue
			}
		}
		outGames = append(outGames, game)
	}
	return outGames
}

func initTwoCardTypeMap() map[int32]int32 {
	m := make(map[int32]int32)
	for _, value := range enumgame.TwoCardCardType {
		m[value] = 0
	}
	return m
}

func initUserCard(game *mdgame.Twocard, goLua *lua.LState) error {

	if err := goLua.DoString("return G_Reset()"); err != nil {
		log.Err("two card G_Reset %+v", err)
		return errgame.ErrGoLua
	}
	for _, ui := range game.GameResult.List {
		if err := goLua.DoString("return G_GetCards()"); err != nil {
			log.Err("two card G_GetCards err %v", err)
			return errgame.ErrGoLua
		}
		getCards := goLua.Get(-1)
		goLua.Pop(1)
		var cardList []string
		if cardsMap, ok := getCards.(*lua.LTable); ok {
			cardsMap.ForEach(func(key lua.LValue, value lua.LValue) {
				if cards, ok := value.(*lua.LTable); ok {
					var cardType string
					var cardValue string
					cards.ForEach(func(k lua.LValue,
						v lua.LValue) {
						if k.String() == "_type" {
							cardType = v.String()
						} else if k.String() ==
							"_value" {
							cardValue = v.String()
						}
					})
					cardList = append(cardList,
						cardType+"_"+cardValue)
				} else {
					log.Err("two card cardsMap value err %v",
						value)
				}
			})
			if len(cardList) == 0 {
				log.Err("two card cardList nil err %v", cardsMap)
				return errgame.ErrGoLua
			}
			ui.CardList = cardList
			ui.TotalScore = 0
		} else {
			log.Err("two card cardsMap err %v", cardsMap)
			return errgame.ErrGoLua
		}
	}
	return nil
}

func autoSetBankerScore(game *mdgame.Twocard) {
	for _, userResult := range game.GameResult.List {
		if userResult.Status == enumgame.UserStatusInit {
			userResult.Bet = 1
			userResult.Status = enumgame.UserStatusSetBet
		}
	}
}

func SetBet(uid int32, key int32, mdr *mdroom.Room) error {
	var value int32
	if key < 1 || key > 5 {
		return errgame.ErrParam
	}
	value = key
	//if v, ok := enumgame.BetScoreMap[key]; !ok {
	//	return errgame.ErrParam
	//} else {
	//	value = v
	//}

	if mdr.Status > enumroom.RoomStatusStarted {
		return errroom.ErrGameIsDone
	}
	if mdr.Giveup == enumroom.WaitGiveUp {
		return errroom.ErrInGiveUp
	}
	game, err := cachegame.GetGame(mdr.RoomID)
	if err != nil {
		return err
	}
	if game == nil {
		return errgame.ErrGameNotExist
	}
	if game.Status != enumgame.GameStatusInit {
		return errgame.ErrBetDone
	}

	allReady, userResult := getUserAndAllOtherStatusReady(game, uid,
		enumgame.GetBetStatus)
	if userResult == nil {
		return errgame.ErrUserNotInGame
	}

	if userResult.Role == enumroom.UserRoleMaster {
		return errgame.ErrBankerNoBet
	}

	if userResult.Status > enumgame.UserStatusSetBet {
		return errgame.ErrAlreadySetBet
	}

	userResult.Status = enumgame.UserStatusSetBet
	userResult.Bet = value
	if allReady {
		game.Status = enumgame.GameStatusAllBet
	}

	err = cachegame.UpdateGame(game)
	if err != nil {
		log.Err("two card set session failed, %v", err)
		return err
	}

	return nil //
}

func autoSubmitCard(game *mdgame.Twocard) error {
	for _, userResult := range game.GameResult.List {
		if userResult.Status == enumgame.UserStatusSetBet {
			sort.Strings(userResult.CardList)
			mdHeadCard := &mdgame.UserCard{
				CardList: userResult.CardList,
			}
			userResult.Cards = mdHeadCard
			userResult.Status = enumgame.UserStatusSubmitCard
		}
	}
	game.Status = enumgame.GameStatusAllSubmitCard
	err := cachegame.UpdateGame(game)
	if err != nil {
		log.Err("two card set session failed, %v", err)
		return err
	}
	return nil
}

func SubmitCard(uid int32, room *mdroom.Room) (*mdgame.Twocard, error) {
	if room.Status > enumroom.RoomStatusStarted {
		if room.Giveup == enumroom.WaitGiveUp {
			return nil, errroom.ErrInGiveUp
		}
		return nil, errroom.ErrGameIsDone
	}

	game, err := cachegame.GetGame(room.RoomID)
	if game == nil {
		return nil, errgame.ErrGameNoFind
	}
	if game.Status != enumgame.GameStatusSubmitCard {
		return nil, errgame.ErrSubmitCardDone
	}

	allReady, userResult := getUserAndAllOtherStatusReady(game, uid,
		enumgame.GetSubmitCardStatus)

	if userResult == nil {
		return nil, errgame.ErrUserNotInGame
	}
	fmt.Printf("SubmitCard:%d\n",userResult.Status)
	if userResult.Status > enumgame.UserStatusSubmitCard {
		return nil, errgame.ErrAlreadySubmitCard
	}

	sort.Strings(userResult.CardList)
	mdHeadCard := &mdgame.UserCard{
		CardList: userResult.CardList,
	}
	userResult.Cards = mdHeadCard
	userResult.Status = enumgame.UserStatusSubmitCard

	if allReady {
		game.Status = enumgame.GameStatusAllSubmitCard
	}

	err = cachegame.UpdateGame(game)
	if err != nil {
		log.Err("two card set session failed, %v", err)
		return nil, err
	}
	return game, nil //
}

func getUserAndAllOtherStatusReady(game *mdgame.Twocard, uid int32,
	getType int32) (bool, *mdgame.UserInfo) {
	var userResult *mdgame.UserInfo
	allReady := true
	for _, user := range game.GameResult.List {
		if user.UserID == uid {
			userResult = user
			if getType == enumgame.GetBetStatus &&
				userResult.Role == enumroom.UserRoleMaster {
				return false, userResult
			}
		} else if getType == enumgame.GetBetStatus {
			if user.Role == enumroom.UserRoleMaster {
				user.Status = enumgame.UserStatusSetBet
			}
			if user.Status != enumgame.UserStatusSetBet {
				allReady = false
			}
		} else if getType == enumgame.GetSubmitCardStatus {
			if user.Status != enumgame.UserStatusSubmitCard {
				allReady = false
			}
		}
	}
	return allReady, userResult
}

func GameResultList(rid int32) (*pbtwo.GameResultListReply, error) {
	var list []*pbtwo.GameResult
	games, err := dbgame.GetTwoCardByRoomID(db.DB(), rid)
	if err != nil {
		return nil, err
	}
	for _, game := range games {
		list = append(list, game.ToProto())
	}
	out := &pbtwo.GameResultListReply{
		List: list,
	}
	return out, nil
}

func randomUserDice(game *mdgame.Twocard) {
	ud := &mdgame.UserDice{}
	ud.DiceAPoints = rand.Int31n(5) + 1
	ud.DiceBPoints = rand.Int31n(5) + 1
	userIndex := (ud.DiceAPoints + ud.DiceBPoints) % int32(len(game.Ids))
	ud.UserID = game.Ids[userIndex]
	game.GameResult.UserDice = ud
}

func GameRecovery(rid int32) (*mdgame.Twocard, error) {
	game, err := cachegame.GetGame(rid)
	if err != nil {
		return nil, err
	}
	if game == nil {
		game, err = dbgame.GetLastTwoCardByRoomID(db.DB(), rid)
		if err != nil {
			return nil, err
		}
	}
	if game == nil {
		return nil, errgame.ErrGameNotExist
	}
	return game, nil
}

func GameExist(uid int32, rid int32) (*pbtwo.RecoveryReply, error) {
	out := &pbtwo.RecoveryReply{}
	_, roomRecovery, err := room.CheckRoomExist(uid, rid)
	if err != nil {
		return nil, err
	}
	out.RoomExist = roomRecovery.ToProto()
	out.RoomExist.Room.CreateOrEnter = enumroom.EnterRoom
	out.RoomExist.Room.OwnerID = out.RoomExist.Room.UserList[0].UserID
	if roomRecovery.Status < enumroom.RecoveryGameStart && roomRecovery.Status != enumroom.RecoveryInitNoReady {
		return out, nil
	}
	game, err := GameRecovery(roomRecovery.Room.RoomID)
	if err != nil {
		return nil, err
	}
	out.TwoCardExist = game.ToProto()
	for _, gr := range out.TwoCardExist.List {
		if gr.UserID != uid && game.Status < enumgame.GameStatusDone {
			gr.CardList = nil
			gr.Cards = nil
		}
	}
	var time int32
	switch game.Status {
	case enumgame.GameStatusInit:
		time = enumgame.SetBetTime
		break
	case enumgame.GameStatusAllSubmitCard:
		time = enumgame.SubmitCardTime
		break
	}
	out.CountDown = &pbtwo.CountDown{
		ServerTime: game.OpDateAt.Unix(),
		Count:      time,
	}
	return out, nil
}

func CleanGame() error {
	var gids []int32
	rids, err := cacheroom.GetAllDeleteRoomKey(enumroom.TwoCardGameType)
	if err != nil {
		log.Err("get two card clean room err:%v", err)
		return err
	}
	for _, rid := range rids {
		game, err := cachegame.GetGame(rid)
		if err != nil {
			log.Err("get two card give up room err:%d|%v", rid, err)
			continue
		}
		if game != nil {
			log.Debug("clean two card game:%d|%d|%+v\n", game.GameID, game.RoomID, game.Ids)
			gids = append(gids, game.GameID)
			err = cachegame.DeleteGame(game)
			if err != nil {
				log.Err(" delete two card set session failed, %v",
					err)
				continue
			}
			err = cacheroom.CleanDeleteRoom(enumgame.GameID, game.RoomID)
			if err != nil {
				log.Err(" delete two card delete room session failed,roomid:%d,err: %v", game.RoomID,
					err)
				continue
			}
		} else {
			err = cacheroom.CleanDeleteRoom(enumgame.GameID, rid)
			if err != nil {
				log.Err(" delete null game two card delete room session failed,roomid:%d,err: %v", rid,
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
