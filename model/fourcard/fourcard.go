package fourcard

import (
	"encoding/json"
	"fmt"
	cacheroom "playcards/model/room/cache"
	dbroom "playcards/model/room/db"
	enumroom "playcards/model/room/enum"
	errroom"playcards/model/room/errors"
	mdroom "playcards/model/room/mod"
	cachegame "playcards/model/fourcard/cache"
	dbgame "playcards/model/fourcard/db"
	enumgame "playcards/model/fourcard/enum"
	errgame "playcards/model/fourcard/errors"
	mdgame "playcards/model/fourcard/mod"
	"math/rand"
	"playcards/utils/db"
	"playcards/utils/log"
	"github.com/jinzhu/gorm"
	"github.com/yuin/gopher-lua"
	"time"
)

//var GoLua *lua.LState
//
//func InitGoLua(gl *lua.LState) {
//	GoLua = gl
//}

func CreateGame(goLua *lua.LState) []*mdgame.FourCard {
	mdrooms := cacheroom.GetAllRoomByGameTypeAndStatus(enumroom.FourCardGameType, enumroom.RoomStatusAllReady)
	if mdrooms == nil && len(mdrooms) == 0 {
		return nil
	}
	if len(mdrooms) == 0 {
		return nil
	}
	var newGames []*mdgame.FourCard

	for _, mdroom := range mdrooms {
		var newGameResult *mdgame.GameResult
		var userResults []*mdroom.GameUserResult
		var bankerID int32
		for _, user := range mdroom.Users {
			if mdroom.RoundNow == 1 {
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
			if user.Role == enumroom.UserRoleMaster {
				bankerID = user.UserID
			}
			ui := &mdgame.UserInfo{
				UserID:     user.UserID,
				Status:     enumgame.UserStatusInit,
				Bet:        0,
				Role:       user.Role,
				TotalScore: 0,
			}
			newGameResult.List = append(newGameResult.List, ui)
		}
		if mdroom.RoundNow == 1 {
			mdroom.UserResults = userResults
		}
		now := gorm.NowFunc()
		game := &mdgame.FourCard{
			RoomID:     mdroom.RoomID,
			BankerID:   bankerID,
			Status:     enumgame.GameStatusInit,
			Index:      mdroom.RoundNow,
			PassWord:   mdroom.Password,
			GameResult: newGameResult,
			OpDateAt:   &now,
			Ids:        mdroom.Ids,
		}
		mdroom.Status = enumroom.RoomStatusStarted
		f := func(tx *gorm.DB) error {
			if mdroom.RoundNow == 1 {
				//if mdroom.RoomType != enumroom.RoomTypeClub && mdroom.Cost != 0 {
				//	err := bill.GainGameBalance(mdroom.PayerID, mdroom.RoomID, enumbill.JournalTypeThirteen,
				//		enumbill.JournalTypeThirteenUnFreeze, &mdbill.Balance{Amount: -mdroom.Cost, CoinType: mdroom.CostType})
				//	if err != nil {
				//		log.Err("four card create user balance failed,%v | %+v", mdroom.PayerID, err)
				//		return err
				//	}
				//}
				for _, user := range mdroom.Users {
					pr := &mdroom.PlayerRoom{
						UserID:    user.UserID,
						RoomID:    mdroom.RoomID,
						GameType:  mdroom.GameType,
						PlayTimes: 0,
					}
					dbroom.CreatePlayerRoom(tx, pr)
				}
			}

			_, err := dbroom.UpdateRoom(tx, mdroom)
			if err != nil {
				log.Err("four card room update set session failed, roomid:%d,err:%+v", mdroom.RoomID, err)
				return err
			}

			err = dbgame.CreateGame(tx, game)
			if err != nil {
				log.Err("four card create set session failed,roomid:%d, err:%+v", mdroom.RoomID, err)
				return err
			}

			err = cacheroom.UpdateRoom(mdroom)
			if err != nil {
				log.Err("room update set session failed,roomid:%d,err: %+v", mdroom.RoomID, err)
				return err
			}

			err = cachegame.SetGame(game)
			if err != nil {
				log.Err("four card create set redis failed,%v | %+v", mdroom, err)
				return err
			}
			return nil
		}
		//go db.Transaction(f)

		err := db.Transaction(f)
		if err != nil {
			log.Err("four card create failed,%v | %+v", game, err)
			continue
		}
		newGames = append(newGames, game)
	}
	return newGames
}

//游戏结算逻辑
func UpdateGame(goLua *lua.LState) []*mdgame.FourCard {
	games, err := cachegame.GetAllGameByStatus(enumgame.GameStatusDone)
	if err != nil {
		log.Err("four card get all game by status failed, %v", err)
		return nil
	}
	if len(games) == 0 {
		return nil
	}
	//游戏结算结果集合
	var outGames []*mdgame.FourCard
	for _, game := range games {
		if game.Status == enumgame.GameStatusInit {
			sub := time.Now().Sub(*game.OpDateAt)
			if sub.Seconds() > enumgame.SetBetTime {
				autoSetBankerScore(game)
				game.Status = enumgame.GameStatusAllBet
				err := cachegame.UpdateGame(game)
				if err != nil {
					log.Err("niuniu set session failed, %v", err)
					continue
				}
			}
		} else if game.Status == enumgame.GameStatusAllBet {
			randomUserDice(game)
			initUserCard(game, goLua)
			game.Status = enumgame.GameStatusOrdered
		} else if game.Status == enumgame.GameStatusOrdered {

		} else if game.Status == enumgame.GameStatusAllSubmitCard {
			room, err := cacheroom.GetRoom(game.PassWord)
			if err != nil {
				//print(err)
				log.Err("four card room get session failed, roomid:%d,pwd:%s,err:%v", game.RoomID, game.PassWord, err)
				continue
			}
			if room == nil {
				log.Err("four card room get session nil, %v|%d", game.PassWord, game.RoomID)
				continue
			}

			game.MarshalGameResult()
			//room.MarshalGameUserResult()
			if err := goLua.DoString(fmt.
			Sprintf("return G_CalculateRes('%s','%s')",
				game.GameResult, room.GameParam)); err != nil {
				log.Err("four card G_CalculateRes %v|\n%v|\n%v\n",
					game.GameResult, room.GameParam, err)
				continue
			}

			luaResult := goLua.Get(-1)
			goLua.Pop(1)
			var results *mdgame.GameResult
			if err := json.Unmarshal([]byte(luaResult.String()),
				&results); err != nil {
				log.Err("four card lua str do struct %v", err)
			}
			game.GameResult = results
			for _, result := range game.GameResult.List {
				m := initFourCardTypeMap()
				for _, userResult := range room.UserResults {
					if userResult.UserID == result.UserID {
						if len(userResult.GameCardCount) > 0 {
							if err := json.Unmarshal([]byte(userResult.GameCardCount), &m); err != nil {
								log.Err("four card lua str do struct %v", err)
							}
						}

						if _, ok := m[result.HeadCards.CardType]; ok {
							m[result.HeadCards.CardType]++
						}
						if _, ok := m[result.TailCards.CardType]; ok {
							m[result.TailCards.CardType]++
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
					}
				}
			}
			game.Status = enumgame.GameStatusDone
			room.Status = enumroom.RoomStatusReInit
			f := func(tx *gorm.DB) error {
				game, err = dbgame.UpdateGame(tx, game)
				if err != nil {
					log.Err("four card update db failed, %v|%v", game, err)
					return err
				}
				room, err = dbroom.UpdateRoom(tx, room)
				if err != nil {
					log.Err("four card update room db failed, %v|%v", game, err)
					return err
				}
				return nil
			}
			err = db.Transaction(f)
			if err != nil {
				log.Err("four card update failed, %v", err)
				continue
			}
			err = cachegame.DeleteGame(game)
			if err != nil {
				log.Err("four card room del session failed, roomid:%d,pwd:%s,err:%v", game.RoomID, game.PassWord, err)
				continue
			}

			err = cacheroom.UpdateRoom(room)
			if err != nil {
				log.Err("four card room update room redis failed,%v | %v",
					room, err)
				continue
			}
			//UpdateNiuniu(niuniu, false)
		}
		outGames = append(outGames, game)
	}
	return outGames
}

func initFourCardTypeMap() map[string]int32 {
	m := make(map[string]int32)
	for _, value := range enumgame.FourCardCardType {
		m[value] = 0
	}
	return m
}

func initUserCard(game *mdgame.FourCard, goLua *lua.LState) error {
	if err := goLua.DoString("return G_Reset()"); err != nil {
		log.Err("four card G_Reset %+v", err)
		return errgame.ErrGoLua
	}
	for _, ui := range game.GameResult.List {
		if err := goLua.DoString("return G_GetCards()"); err != nil {
			log.Err("four card G_GetCards err %v", err)
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
					log.Err("four card cardsMap value err %v",
						value)
				}
			})
			if len(cardList) == 0 {
				log.Err("four card cardList nil err %v", cardsMap)
				return errgame.ErrGoLua
			}
			ui.CardList = cardList
		} else {
			log.Err("four card cardsMap err %v", cardsMap)
			return errgame.ErrGoLua
		}
	}
	return nil
}

func autoSetBankerScore(game *mdgame.FourCard) {
	for _, userResult := range game.GameResult.List {
		if userResult.Status == enumgame.UserStatusInit {
			userResult.Bet = 0
			userResult.Status = enumgame.UserStatusSetBet
		}
	}
}

func SetBet(uid int32, key int32, mdr *mdroom.Room) ([]int32, error) {
	var value int32
	if v, ok := enumgame.BetScoreMap[key]; !ok {
		return nil, errgame.ErrParam
	} else {
		value = v
	}

	if mdr.Status > enumroom.RoomStatusStarted {
		return nil, errroom.ErrGameIsDone
	}
	if mdr.Giveup == enumroom.WaitGiveUp {
		return nil, errroom.ErrInGiveUp
	}
	game, err := cachegame.GetGame(mdr.RoomID)

	if game.Status != enumgame.GameStatusInit {
		return nil, errgame.ErrBetDone
	}

	allReady, userResult := getUserAndAllOtherStatusReady(game, uid,
		enumgame.GetBetStatus)

	if userResult == nil {
		return nil, errgame.ErrUserNotInGame
	}

	if userResult.Role == enumroom.UserRoleMaster {
		return nil, errgame.ErrBankerNoBet
	}

	if userResult.Status > enumgame.UserStatusSetBet {
		return nil, errgame.ErrAlreadySetBet
	}

	userResult.Status = enumgame.UserStatusSetBet
	userResult.Bet = value
	if allReady {
		game.Status = enumgame.GameStatusAllBet
	}

	err = cachegame.UpdateGame(game)
	if err != nil {
		log.Err("four card set session failed, %v", err)
		return nil, err
	}

	return game.Ids, nil //
}

func SubmitCard(uid int32, room *mdroom.Room, head []string, tail []string) ([]int32, error) {
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
	if game.Status != enumgame.GameStatusOrdered {
		return nil, errgame.ErrSubmitCardDone
	}

	if cacheroom.GetRoomTestConfigKey("FourCardCheckHasCards") != "0" {
		var checkCards []string
		for _, userTmp := range game.GameResult.List {
			if userTmp.UserID == uid {
				checkCards = userTmp.CardList
			}
		}
		checkList := append(head, tail...)
		checkHasCard := checkHasCards(checkList, checkCards)
		if !checkHasCard {
			return nil, errgame.ErrCardNotExist
		}
	}

	allReady, userResult := getUserAndAllOtherStatusReady(game, uid,
		enumgame.GetSubmitCardStatus)

	if userResult == nil {
		return nil, errgame.ErrUserNotInGame
	}

	if userResult.Status > enumgame.UserStatusSubmitCard {
		return nil, errgame.ErrAlreadySubmitCard
	}
	userResult.HeadCards.CardList = head
	userResult.TailCards.CardList = tail
	userResult.Status = enumgame.UserStatusSubmitCard

	if allReady {
		game.Status = enumgame.GameStatusAllSubmitCard
	}

	err = cachegame.UpdateGame(game)
	if err != nil {
		log.Err("four card set session failed, %v", err)
		return nil, err
	}
	return game.Ids, nil //
}

func getUserAndAllOtherStatusReady(game *mdgame.FourCard, uid int32,
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

func randomUserDice(game *mdgame.FourCard) {
	ud := &mdgame.UserDice{}
	ud.DiceAPoints = rand.Int31n(5) + 1
	ud.DiceBPoints = rand.Int31n(5) + 1
	userIndex := (ud.DiceAPoints + ud.DiceBPoints) % int32(len(game.Ids))
	ud.UserID = game.Ids[userIndex]
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
