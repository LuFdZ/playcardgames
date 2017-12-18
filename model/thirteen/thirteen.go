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
	//"playcards/utils/env"
	"playcards/model/bill"
	"playcards/utils/db"
	"playcards/utils/log"
	"strconv"
	"github.com/jinzhu/gorm"
	"github.com/yuin/gopher-lua"
)

var GoLua *lua.LState

func InitGoLua(gl *lua.LState) {
	GoLua = gl
}

func CreateThirteen() []*mdt.Thirteen {
	//rooms, err := GetRoomsByStatusAndGameType()
	//if err != nil {
	//	log.Err("get rooms by status and game type err :%+v", err)
	//	return nil
	//}

	f := func(r *mdr.Room) bool {
		if r.Status == enumr.RoomStatusAllReady && r.GameType == enumt.GameID {
			return true
		}
		return false
	}
	rooms := cacher.GetAllRoom(f)
	if rooms == nil && len(rooms) == 0 {
		return nil
	}

	if len(rooms) == 0 {
		return nil
	}
	var newGames []*mdt.Thirteen

	for _, room := range rooms {
		var userResults []*mdr.GameUserResult
		var groupCards []*mdt.GroupCard
		var bankerID int32
		//l := lua.NewState()
		//defer l.Close()
		//filePath := env.GetCurrentDirectory() + "/lua/thirteenlua/Logic.lua"
		//if err := l.DoFile(filePath); err != nil {
		//	log.Err("thirteen logic do file %+v", err)
		//	continue
		//}
		//ostimeA := time.Now().UnixNano()
		//ostimeB := ostimeA<<32 | ostimeA>>32
		//fmt.Printf("ostime:%d|%d\n",ostimeA,ostimeB)
		//log.Err("create thirteen seed:%d|%d\n", ostimeA, ostimeB)
		//if err := GoLuaCreate.DoString(fmt.Sprintf("return Logic:new(%d)", ostimeB)); err != nil {
		//	log.Err("thirteen logic do string %+v", err)
		//	continue
		//}
		//if err := l.DoString("return Logic:new()"); err != nil {
		//	log.Err("thirteen logic do string %+v", err)
		//	continue
		//}
		//logic := l.Get(1)

		if err := GoLua.DoString("return G_Reset()"); err != nil {
			log.Err("thirteen G_Reset %+v", err)
			continue
		}
		//GoLuaCreate.Pop(1)

		for _, user := range room.Users {
			if room.RoundNow == 1 {
				userResult := &mdr.GameUserResult{
					UserID:   user.UserID,
					Nickname: user.Nickname,
					Role:     user.Role,
					Win:      0,
					Lost:     0,
					Tie:      0,
					Score:    0,
				}
				userResults = append(userResults, userResult)
			}
			if err := GoLua.DoString("return G_GetCards()"); err != nil {
				log.Err("thirteen G_GetCards %+v", err)
			}
			getCards := GoLua.Get(-1)
			GoLua.Pop(1)
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

		if room.RoundNow == 1 {
			room.UserResults = userResults
		}
		thirteen := &mdt.Thirteen{
			RoomID:   room.RoomID,
			BankerID: bankerID,
			Status:   enumt.GameStatusInit,
			Index:    room.RoundNow,
			//GameLua: l,
			Cards: groupCards,
			Ids:   room.Ids,
		}
		room.Status = enumr.RoomStatusStarted
		f := func(tx *gorm.DB) error {
			if room.RoundNow == 1 {
				if room.RoomType != enumr.RoomTypeClub && room.Cost != 0 {
					err := bill.GainGameBalance(room.PayerID, room.RoomID, enumbill.JournalTypeThirteen,
						enumbill.JournalTypeThirteenUnFreeze, &mdbill.Balance{Amount: -room.Cost, CoinType: room.CostType})
					if err != nil {
						log.Err("thirteen create user balance failed,%v | %+v", room.PayerID, err)
						return err
					}
				}

				for _, user := range room.Users {
					pr := &mdr.PlayerRoom{
						UserID:    user.UserID,
						RoomID:    room.RoomID,
						GameType:  room.GameType,
						PlayTimes: 0,
					}
					dbr.CreatePlayerRoom(tx, pr)
				}
			}

			_, err := dbr.UpdateRoom(tx, room)
			if err != nil {
				log.Err("thirten room update set session failed, roomid:%d,err:%+v", room.RoomID, err)
				return err
			}

			err = dbt.CreateThirteen(tx, thirteen)
			if err != nil {
				log.Err("thirteen create set session failed,roomid:%d, err:%+v", room.RoomID, err)
				return err
			}

			err = cacher.UpdateRoom(room)
			if err != nil {
				log.Err("room update set session failed,roomid:%d,err: %+v", room.RoomID, err)
				return err
			}

			err = cachet.SetGame(thirteen, room.MaxNumber, room.Password)
			if err != nil {
				log.Err("thirteen create set redis failed,%v | %+v", room, err)
				return err
			}

			for _, user := range room.Users {
				err = cachet.SetGameUser(room.RoomID, user.UserID)
				if err != nil {
					log.Err("thirteen create set game user redis failed,%v | %+v", room, err)
					return err
				}
			}

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
func UpdateGame() []*mdt.Thirteen { //[]*mdt.GameResultList
	//从数据库获取未结算的游戏信息
	//thirteens, err := GetThirteenByStatusAndGameType()
	//if err != nil {
	//	print(err)
	//	log.Err("get thirteen by status and game type err :%v", err)
	//	return nil
	//}

	f := func(r *mdt.Thirteen) bool {
		if r.Status == enumt.GameStatusStarted {
			return true
		}
		return false
	}
	thirteens := cachet.GetAllThirteen(f)
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
		pwd := cachet.GetRoomPaawordRoomID(thirteen.RoomID)
		room, err := cacher.GetRoom(pwd)
		if err != nil {
			log.Err("room get session failed, %v", err)
			continue
		}
		if room == nil {
			log.Err("room get session failed, %v", err)
			continue
		}

		var results []*mdt.ThirteenResult
		//l := lua.NewState()
		//defer l.Close()
		//filePath := env.GetCurrentDirectory() + "/lua/thirteenlua/Logic.lua"
		//if err := l.DoFile(filePath); err != nil {
		//	log.Err("thirteen clean logic do file %v", err)
		//	continue
		//}
		//if err := GoLuaUpdate.DoString(fmt.Sprintf("return Logic:new(%d)", 0)); err != nil {
		//	log.Err("thirteen logic do string %+v", err)
		//	continue
		//}
		thirteen.MarshalUserSubmitCards()
		//fmt.Printf("GameParam:%v\n%v\n",thirteen.UserSubmitCards,room.GameParam)
		if err := GoLua.DoString(fmt.Sprintf("return G_GetResult('%s','%s')",
			thirteen.UserSubmitCards, room.GameParam)); err != nil {
			log.Err("thirteen G_GetResult %v", err)
			continue
		}

		luaResult := GoLua.Get(-1)
		if err := json.Unmarshal([]byte(luaResult.String()), &results); err != nil {
			log.Err("thirteen lua str do struct %v", err)
			continue
		}

		resultList.Result = results

		var resultArray []*mdr.GameUserResult
		for _, result := range resultList.Result {
			for _, userResult := range room.UserResults {
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
		room.Status = enumr.RoomStatusReInit

		var roomparam *mdr.ThirteenRoomParam
		if err := json.Unmarshal([]byte(room.GameParam), &roomparam); err != nil {
			log.Err("thirteen clean unmarshal room param failed, %v", err)
			continue
		}

		if roomparam.BankerAddScore > 0 && roomparam.Times != enumt.TimesDefault {
			for i := 0; i < len(room.Users); i++ {
				//room.Users[i].Ready = enumr.UserUnready
				//十三张一局结束后 轮庄
				if room.Users[i].Role == enumr.UserRoleMaster {
					if bankerScore <= 0 {
						if i == len(room.Users)-1 {
							room.Users[0].Role = enumr.UserRoleMaster
						} else {
							room.Users[i+1].Role = enumr.UserRoleMaster
						}
						room.Users[i].Role = enumr.UserRoleSlave
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
			r, err := dbr.UpdateRoom(tx, room)
			if err != nil {
				return err
			}
			room = r
			return nil
		}
		//go db.Transaction(f)
		err = db.Transaction(f)
		if err != nil {
			//print(err)
			log.Err("thirteen update failed, %v", err)
			return nil
		}
		err = cachet.DeleteGame(thirteen.RoomID)
		if err != nil {
			log.Err("thirteen set session failed, %v", err)
			return nil
		}
		err = cacher.UpdateRoom(room)
		if err != nil {
			log.Err("room update room redis failed,%v | %v", room, err)
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

func SubmitCard(uid int32, submitCard *mdt.SubmitCard, room *mdr.Room) ([]int32, error) {
	if room.Status > enumr.RoomStatusStarted {
		return nil, errors.ErrGameIsDone
	}
	if room.Giveup == enumr.WaitGiveUp {
		return nil, errors.ErrInGiveUp
	}
	isReady := cachet.IsGamePlayerReady(room.RoomID, uid)

	if isReady == 0 {
		return nil, errorst.ErrUserNotInGame
	} else if isReady == 2 {
		return nil, errorst.ErrAlreadySubmitCard
	}

	thirteen, err := cachet.GetGame(room.RoomID)
	if err != nil {
		return nil, err
	}

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

	for _, user := range room.Users {
		if user.UserID == uid {
			submitCard.Role = user.Role
		}
	}

	submitCard.UserID = uid
	thirteen.SubmitCards = append(thirteen.SubmitCards, submitCard)
	if thirteen.Status > enumt.GameStatusInit {
		return nil, errors.ErrGameIsDone
	}
	playerNow := cachet.GetGamePlayerNowRoomID(room.RoomID)
	playerNow += 1

	if playerNow == room.MaxNumber {
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

	err = cachet.UpdateGameUser(thirteen, uid, playerNow)
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

	rids, err := cacher.GetAllDeleteRoomKey(enumt.GameID)
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
			err = cachet.DeleteGame(rid)
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
