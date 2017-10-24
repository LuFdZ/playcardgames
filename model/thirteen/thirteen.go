package thirteen

import (
	"encoding/json"
	"fmt"
	dbbill "playcards/model/bill/db"
	enumbill "playcards/model/bill/enum"
	mdbill "playcards/model/bill/mod"
	cacher "playcards/model/room/cache"
	dbr "playcards/model/room/db"
	enumr "playcards/model/room/enum"
	errors "playcards/model/room/errors"
	mdr "playcards/model/room/mod"
	cachet "playcards/model/thirteen/cache"
	dbt "playcards/model/thirteen/db"
	enumt "playcards/model/thirteen/enum"
	errorst "playcards/model/thirteen/errors"
	mdt "playcards/model/thirteen/mod"
	pbt "playcards/proto/thirteen"
	"playcards/utils/db"
	"playcards/utils/log"
	"strconv"

	"github.com/jinzhu/gorm"
	"github.com/yuin/gopher-lua"
)

//游戏结算逻辑
func CleanGame() []*mdt.GameResultList {
	//从数据库获取未结算的游戏信息
	thirteens, err := GetThirteenByStatusAndGameType()
	if err != nil {
		print(err)
		log.Err("get thirteen by status and game type err :%v", err)
		return nil
	}
	if len(thirteens) == 0 {
		return nil
	}
	//游戏结算结果集合
	var resultListArray []*mdt.GameResultList

	for _, thirteen := range thirteens {
		resultList := &mdt.GameResultList{}
		resultList.RoomID = thirteen.RoomID
		thirteen.Status = enumt.GameStatusDone
		var bankerScore int32
		//var resultArray []*mdt.GameResult

		//获取游戏所属房间缓存 更新房间信息
		pwd := cachet.GetRoomPaawordRoomID(thirteen.RoomID)
		room, err := cacher.GetRoom(pwd)
		if err != nil {
			print(err)
			log.Err("room get session failed, %v", err)
			continue
		}
		if room == nil {
			print(err)
			log.Err("room get session failed, %v", err)
			continue
		}

		var results []*mdt.ThirteenResult
		l := lua.NewState()
		defer l.Close()
		if err := l.DoFile("lua/thirteenlua/Logic.lua"); err != nil {
			log.Err("thirteen clean logic do file %v", err)
			continue
		}

		if err := l.DoString("return Logic:new()"); err != nil {
			log.Err("thirteen get result do string %v", err)
		}

		if err := l.DoString(fmt.Sprintf("return Logic:GetResult('%s','%s')",
			thirteen.UserSubmitCards, room.GameParam)); err != nil {
			//fmt.Printf("clean result CCCCCC %v", err)
			log.Err("thirteen get result do string %v", err)
			continue
		}

		luaResult := l.Get(-1)
		//fmt.Printf("AAAAA thirteen lua result %v: \n", luaResult)
		if err := json.Unmarshal([]byte(luaResult.String()), &results); err != nil {
			//fmt.Printf("thirteen set lua str do struct %v: \n", err)
			log.Err("thirteen lua str do struct %v", err)
			continue
		}

		resultList.Result = results

		var resultArray []*mdr.GameUserResult
		for _, result := range resultList.Result {
			//fmt.Printf("thirteen AAAAAAAAAAA:%d \n", result.UserID)
			for _, userResult := range room.UserResults {
				m := InitThirteenGameTypeMap()
				//fmt.Printf("Thirteen Result Map:%+v\n", m)
				if userResult.UserID == result.UserID {
					if len(userResult.GameCardCount) > 0 {
						if err := json.Unmarshal([]byte(userResult.GameCardCount), &m); err != nil {
							//fmt.Printf("room param str to map err %v: \n", err)
							log.Err("thirteen game card count lua str do struct %v", err)
						}
					}
					ts, _ := strconv.ParseInt(result.Settle.TotalScore, 10, 32)
					userResult.Score += int32(ts)

					if ts > 0 {
						userResult.Win += 1
					} else if ts == 0 {
						userResult.Tie += 1
					} else if ts == 0 {
						userResult.Lost += 1
					}
					//fmt.Printf("AAAThirteen Result Head:%s\n", result.Result.Head.GroupType)
					if _, ok := m[result.Result.Head.GroupType]; ok {
						m[result.Result.Head.GroupType]++
					}

					//fmt.Printf("AAAThirteen Result Middle:%s\n", result.Result.Middle.GroupType)
					if _, ok := m[result.Result.Middle.GroupType]; ok {
						m[result.Result.Middle.GroupType]++
					}
					//fmt.Printf("AAAThirteen Result Tail:%s\n", result.Result.Tail.GroupType)
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
					//fmt.Printf("Clean Game userResult Array:%+v\n", userResult)
					if userResult.Role == enumr.UserRoleMaster {
						bankerScore = int32(ts)
					}
				}
			}
		}

		resultListArray = append(resultListArray, resultList)
		thirteen.Result = resultList
		room.Status = enumr.RoomStatusReInit

		var roomparam *mdt.ThirteenRoomParam
		if err := json.Unmarshal([]byte(room.GameParam), &roomparam); err != nil {
			//fmt.Printf("BBB json unmarshal err :%v", err)
			log.Err("thirteen clean unmarshal room param failed, %v", err)
			continue
		}

		for i := 0; i < len(room.Users); i++ {
			room.Users[i].Ready = enumr.UserUnready
			//十三张一局结束后 轮庄
			if roomparam.BankerAddScore > 0 {
				if room.Users[i].Role == enumr.UserRoleMaster {
					//fmt.Printf("do round banker:%d|%d", room.Users[i].UserID, bankerScore)
					if bankerScore <= 0 {
						if i == len(room.Users)-1 {
							room.Users[0].Role = enumr.UserRoleMaster
						} else {
							room.Users[i+1].Role = enumr.UserRoleMaster
						}
					}

				}
			}
		}

		f := func(tx *gorm.DB) error {
			//fmt.Printf("UpdateThirteen:%+v", thirteen)
			thirteen, err = dbt.UpdateThirteen(tx, thirteen)
			if err != nil {
				log.Err("thirteen update failed, %v", err)
				return err
			}
			r, err := dbr.UpdateRoom(tx, room)
			if err != nil {
				return err
			}
			room = r
			return nil
		}
		err = db.Transaction(f)
		if err != nil {
			print(err)
			log.Err("thirteen update failed, %v", err)
			return nil
		}
		err = cacher.SetRoom(room)
		if err != nil {
			log.Err("room update room redis failed,%v | %v", room, err)
			return nil
		}

		err = cachet.DeleteGame(thirteen.RoomID)
		if err != nil {
			log.Err("thirteen set session failed, %v", err)
			return nil
		}
	}

	return resultListArray
}

func InitThirteenGameTypeMap() map[string]int32 {
	m := make(map[string]int32)
	for _, value := range enumt.GroupTypeName {
		m[value] = 0
	}
	return m
}

func CreateThirteen() []*mdt.Thirteen {
	rooms, err := GetRoomsByStatusAndGameType()
	if err != nil {
		log.Err("get rooms by status and game type err :%+v", err)
		return nil
	}

	if len(rooms) == 0 {
		return nil
	}
	var newGames []*mdt.Thirteen
	var userResults []*mdr.GameUserResult
	for _, room := range rooms {
		l := lua.NewState()
		defer l.Close()
		if err := l.DoFile("lua/thirteenlua/Logic.lua"); err != nil {
			fmt.Printf("thirteen logic do file %+v", err)
			log.Err("thirteen logic do file %+v", err)
			continue
		}

		if err := l.DoString("return Logic:new()"); err != nil {
			fmt.Printf("thirteen logic do string %+v", err)
			log.Err("thirteen logic do string %+v", err)
			continue
		}
		//logic := l.Get(1)
		l.Pop(1)
		//fmt.Printf("return value is : %#v\n", ret)

		var groupCards []*mdt.GroupCard
		var bankerID int32
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
			if err := l.DoString("return Logic:GetCards()"); err != nil {
				fmt.Printf("thirteen logic do string %+v", err)
				log.Err("thirteen logic do string %+v", err)
			}
			getCards := l.Get(1)
			l.Pop(1)
			//fmt.Printf("return value is : %#v\n", getCards)
			var cardList []string
			if cardsMap, ok := getCards.(*lua.LTable); ok {
				cardsMap.ForEach(func(key lua.LValue, value lua.LValue) {
					if cards, ok := value.(*lua.LTable); ok {
						var cardType string
						var cardValue string
						cards.ForEach(func(k lua.LValue, v lua.LValue) {
							//value, _ := strconv.ParseInt(v.String(), 10, 32)
							if k.String() == "_type" {
								cardType = v.String()
							} else {
								cardValue = v.String()
							}
						})
						// card := mdt.Card{
						// 	Type:  int32(cardType),
						// 	Value: int32(cardValue),
						// }

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

		thirteen := &mdt.Thirteen{
			RoomID:   room.RoomID,
			BankerID: bankerID,
			Status:   enumt.GameStatusInit,
			Index:    room.RoundNow,
			//GameLua: l,
			Cards: groupCards,
		}

		f := func(tx *gorm.DB) error {
			err = dbt.CreateThirteen(tx, thirteen)
			if err != nil {
				fmt.Printf("create thirteen err:%+v", err)
				return err
			}

			room.Status = enumr.RoomStatusStarted
			err = cacher.UpdateRoom(room)
			if err != nil {
				fmt.Printf("room update set session failed, %+v", err)
				log.Err("room update set session failed, %+v", err)
				return err
			}
			_, err = dbr.UpdateRoom(tx, room)

			if room.RoundNow == 1 {
				err := dbbill.GainBalance(tx, room.PayerID,
					&mdbill.Balance{0, 0,
						-int64(room.MaxNumber * room.RoundNumber *
							enumr.ThirteenGameCost)}, //enumt.GameCost
					enumbill.JournalTypeRoom,
					strconv.Itoa(int(room.GameType))+
						room.Password+
						strconv.Itoa(int(room.RoomID)),
					room.Users[0].UserID, enumbill.DefaultChannel)
				if err != nil {
					return err
				}
				room.UserResults = userResults

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

			return nil
		}
		err = db.Transaction(f)
		if err != nil {
			log.Err("thirteen create failed,%v | %+v", thirteen, err)
			continue
		}
		newGames = append(newGames, thirteen)
		cachet.SetGame(thirteen, room.MaxNumber, room.Password)
		if err != nil {
			log.Err("thirteen create set redis failed,%v | %+v", room, err)
			continue
		}
		err = cacher.SetRoom(room)
		//fmt.Printf("GameUserResult : %+v\n", room.UserResults)
		if err != nil {
			log.Err("room create set redis failed,%v | %+v", room, err)
			continue
		}
		for _, user := range room.Users {
			cachet.SetGameUser(room.RoomID, user.UserID)
		}
	}
	return newGames
}

func SubmitCard(uid int32, submitCard *mdt.SubmitCard) (int32, error) {

	pwd := cacher.GetRoomPasswordByUserID(uid)
	if len(pwd) == 0 {
		return 0, errors.ErrUserNotInRoom
	}
	room, err := cacher.GetRoom(pwd)
	if err != nil {
		return 0, err
	}

	if room.Status > enumr.RoomStatusStarted {
		return 0, errors.ErrGameIsDone
	}
	if room.Giveup == enumr.WaitGiveUp {
		return 0, errors.ErrInGiveUp
	}
	isReady := cachet.IsGamePlayerReady(room.RoomID, uid)

	if isReady == 0 {
		return 0, errorst.ErrUserNotInGame
	} else if isReady == 2 {
		return 0, errorst.ErrAlreadySubmitCard
	}

	thirteen, err := cachet.GetGame(room.RoomID)
	if err != nil {
		return 0, err
	}

	// var checkCards []string
	// for _, cardGroup := range thirteen.Cards {
	// 	if cardGroup.UserID == uid {
	// 		checkCards = cardGroup.CardList
	// 	}
	// }

	// checkHasCard := CheckHasCards(submitCard, checkCards)
	// if !checkHasCard {
	// 	return 0, errorst.ErrCardNotExist
	// }

	for _, user := range room.Users {
		if user.UserID == uid {
			submitCard.Role = user.Role
		}
	}

	submitCard.UserID = uid
	thirteen.SubmitCards = append(thirteen.SubmitCards, submitCard)

	if thirteen.Status > enumt.GameStatusInit {
		return 0, errors.ErrGameIsDone
	}
	playerNow := cachet.GetGamePlayerNowRoomID(room.RoomID)
	playerNow += 1

	if playerNow == room.MaxNumber {
		thirteen.Status = enumt.GameStatusStarted
	}

	f := func(tx *gorm.DB) error {
		thirteen, err = dbt.UpdateThirteen(tx, thirteen)
		if err != nil {
			return err
		}
		return nil
	}
	err = db.Transaction(f)
	if err != nil {
		return 0, err
	}

	err = cachet.UpdateGameUser(thirteen, uid, playerNow)
	if err != nil {
		log.Err("thirteen set session failed, %v", err)
		return 0, err
	}
	//fmt.Printf("SubmitCardCCCCCCCCC:%+v /n", thirteen)
	return thirteen.RoomID, nil //

}

func GetRoomsByStatusAndGameType() ([]*mdr.Room, error) {
	var (
		rooms []*mdr.Room
	)
	list, err := dbr.GetRoomsByStatusAndGameType(db.DB(),
		enumr.RoomStatusAllReady, enumt.GameID)
	if err != nil {
		return nil, err
	}
	rooms = list
	return rooms, nil
}

func GetThirteenByStatusAndGameType() ([]*mdt.Thirteen, error) {
	var (
		thirteens []*mdt.Thirteen
	)
	list, err := dbt.GetThirteensByStatus(db.DB(),
		enumt.GameStatusStarted)
	if err != nil {
		return nil, err
	}
	thirteens = list
	return thirteens, nil
}

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

func CleanGiveUpGame() error {
	var gids []int32
	rids, err := dbr.GetGiveUpRoomIDByGameType(db.DB(), enumt.GameID)
	if err != nil {
		log.Err("get thirteen give up room err:%v", err)
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
			if err != nil {
				log.Err(" delete thirteen set session failed, %v", err)
				continue
			}
		}
		fmt.Printf("CleanGiveUpGame:%+v", game)

	}
	if len(gids) > 0 {
		f := func(tx *gorm.DB) error {
			err = dbt.GiveUpGameUpdate(tx, gids)
			if err != nil {
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

func CheckHasCards(submitCards *mdt.SubmitCard, cardList []string) bool {
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
