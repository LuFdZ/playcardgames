package thirteen

import (
	dbbill "playcards/model/bill/db"
	enumbill "playcards/model/bill/enum"
	mdbill "playcards/model/bill/mod"
	cacheroom "playcards/model/room/cache"
	dbr "playcards/model/room/db"
	enumr "playcards/model/room/enum"
	errors "playcards/model/room/errors"
	mdr "playcards/model/room/mod"
	cachethirteen "playcards/model/thirteen/cache"
	dbt "playcards/model/thirteen/db"
	enumt "playcards/model/thirteen/enum"
	errorsthirteen "playcards/model/thirteen/errors"
	mdt "playcards/model/thirteen/mod"
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
		//var result *mdt.GameResult
		var resultArray []*mdt.GameResult

		//获取游戏所属房间缓存 更新房间信息
		pwd := cachethirteen.GetRoomPaawordRoomID(thirteen.RoomID)
		room, err := cacheroom.GetRoom(pwd)
		if err != nil {
			log.Err("room get session failed, %v", err)
			return nil
		}

		//根据玩家牌组计算结果
		for _, cards := range thirteen.SubmitCards {
			result := &mdt.GameResult{}
			result.UserID = cards.UserID
			result.CardsList = cards
			var settleArray []*mdt.Settle
			for _, orderCards := range thirteen.SubmitCards {
				var totalScore int32
				if cards.UserID != orderCards.UserID {
					settle := &mdt.Settle{}
					settle.UserID = orderCards.UserID
					settle.ScoreHead = 100
					settle.ScoreMiddle = 150
					settle.ScoreTail = -50
					settle.Score = 200
					settleArray = append(settleArray, settle)
					totalScore += settle.Score
				}

				//更新房间玩家输赢记录
				for _, result := range room.UserResults {
					if result.UserID == cards.UserID {
						result.Score = totalScore
						if totalScore > 0 {
							result.Win += 1
						} else if totalScore == 0 {
							result.Tie += 1
						} else if totalScore == 0 {
							result.Lost += 1
						}
					}
				}

			}

			result.SettleList = settleArray
			result.CardType = "test"
			resultArray = append(resultArray, result)
		}
		resultList.Results = resultArray
		resultListArray = append(resultListArray, resultList)
		thirteen.Result = resultList
		//房间局数
		//若到最大局数 则房间流程结束 若没到则重置房间状态和玩家准备状态
		// //room.RoundNow += 1
		// if room.RoundNow >= room.RoundNumber {
		// 	room.Status = enumr.RoomStatusDone
		// } else {
		// 	room.Status = enumr.RoomStatusInit
		// 	for _, user := range room.Users {
		// 		user.Ready = enumr.UserUnready
		// 	}
		// }

		room.Status = enumr.RoomStatusReInit

		f := func(tx *gorm.DB) error {
			//fmt.Printf("UpdateThirteen:%+v", thirteen)
			thirteen, err = dbt.UpdateThirteen(tx, thirteen)
			if err != nil {
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
			log.Err("thirteen update failed, %v", err)
			return nil
		}
		// if room.Status == enumr.RoomStatusDone {
		// 	cacheroom.DeleteRoom(room.Password)
		// } else {
		// 	err = cacheroom.SetRoom(room)
		// }
		err = cacheroom.SetRoom(room)
		if err != nil {
			log.Err("room create set redis failed,%v | %v", room, err)
			return nil
		}

		err = cachethirteen.DeleteGame(thirteen.RoomID)
		if err != nil {
			log.Err("thirteen set session failed, %v", err)
			return nil
		}
	}

	return resultListArray
}

func CreateThirteen() []*mdt.Thirteen {
	rooms, err := GetRoomsByStatusAndGameType()
	if err != nil {
		log.Err("get rooms by status and game type err :%v", err)
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
		if err := l.DoFile("lua/Logic.lua"); err != nil {
			log.Err("thirteen logic do file %v", err)
		}

		if err := l.DoString("return Logic:new()"); err != nil {
			log.Err("thirteen logic do string %v", err)
		}
		//logic := l.Get(1)
		l.Pop(1)
		//fmt.Printf("return value is : %#v\n", ret)

		var groupCards []*mdt.GroupCard
		for _, user := range room.Users {
			//for i := 0; i < 4; i++ {
			// var userID int32
			// var role int32
			// if i+1 > len(room.Users) {
			// 	userID = -1
			// 	role = -1
			// } else {
			// 	userID = room.Users[i].UserID
			// 	role = room.Users[i].Role
			// }

			if room.RoundNow == 0 {
				userResult := &mdr.GameUserResult{
					UserID: user.UserID,
					Win:    0,
					Lost:   0,
					Tie:    0,
					Score:  0,
				}
				userResults = append(userResults, userResult)
			}
			if err := l.DoString("return Logic:GetCards()"); err != nil {
				log.Err("thirteen logic do string %v", err)
			}
			getCards := l.Get(1)
			l.Pop(1)

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
						log.Err("thirteen cardsMap value err %v", value)
					}
				})
				groupCard := &mdt.GroupCard{
					UserID:   user.UserID,
					Type:     user.Role,
					Weight:   0,
					CardList: cardList,
				}
				groupCards = append(groupCards, groupCard)
			} else {
				log.Err("thirteen cardsMap err %v", cardsMap)
			}
		}

		thirteen := &mdt.Thirteen{
			RoomID: room.RoomID,
			Status: enumt.GameStatusInit,
			Index:  room.RoundNow,
			//GameLua: l,
			Cards: groupCards,
		}

		f := func(tx *gorm.DB) error {
			if room.RoundNow == 0 {
				err := dbbill.GainBalance(tx, room.Users[0].UserID,
					&mdbill.Balance{0, 0,
						-int64(room.RoundNumber * enumr.ThirteenGameCost / 10)}, //enumt.GameCost
					enumbill.JournalTypeRoom,
					strconv.Itoa(int(room.GameType))+strconv.Itoa(int(room.RoomID)),
					room.Users[0].UserID)

				if err != nil {
					return err
				}
				room.UserResults = userResults
			}

			err = dbt.CreateThirteen(tx, thirteen)
			if err != nil {
				return err
			}

			room.Status = enumr.RoomStatusStarted
			err = cacheroom.UpdateRoom(room)
			if err != nil {
				log.Err("room update set session failed, %v", err)
				return err
			}
			_, err = dbr.UpdateRoom(tx, room)
			return nil
		}
		err = db.Transaction(f)
		if err != nil {
			log.Err("thirteen create failed,%v | %v", thirteen, err)
			continue
		}
		newGames = append(newGames, thirteen)
		cachethirteen.SetGame(thirteen, room.MaxNumber, room.Password)
		if err != nil {
			log.Err("thirteen create set redis failed,%v | %v", room, err)
			continue
		}
		err = cacheroom.SetRoom(room)
		if err != nil {
			log.Err("room create set redis failed,%v | %v", room, err)
			continue
		}
		for _, user := range room.Users {
			cachethirteen.SetGameUser(room.RoomID, user.UserID)
		}
	}
	return newGames
}

func SubmitCard(uid int32, submitCard *mdt.SubmitCard) (int32, error) {

	pwd := cacheroom.GetRoomPasswordByUserID(uid)
	if len(pwd) == 0 {
		return 0, errors.ErrUserNotInRoom
	}
	room, err := cacheroom.GetRoom(pwd)
	if err != nil {
		return 0, err
	}

	if room.Status > enumr.RoomStatusStarted {
		return 0, errors.ErrGameIsDone
	}

	isReady := cachethirteen.IsGamePlayerReady(room.RoomID, uid)

	if isReady == 0 {
		return 0, errorsthirteen.ErrUserNotInGame
	} else if isReady == 2 {
		return 0, errorsthirteen.ErrUserAlready
	}
	//thirteen, err := dbt.GetThitteenByID(db.DB(), gid)

	thirteen, err := cachethirteen.GetGame(room.RoomID)
	if err != nil {
		return 0, err
	}
	//fmt.Printf("SubmitCardAAAAAAAA:%+v /n", thirteen)
	submitCard.UserID = uid
	thirteen.SubmitCards = append(thirteen.SubmitCards, submitCard)

	if thirteen.Status > enumt.GameStatusInit {
		return 0, errors.ErrGameIsDone
	}
	playerNow := cachethirteen.GetGamePlayerNowRoomID(room.RoomID)
	playerNow += 1
	//fmt.Printf("SubmitCardAAAAAAA:%d|%d /n", playerNow, room.MaxNumber)
	if playerNow == room.MaxNumber {
		thirteen.Status = enumt.GameStatusAllReady
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
	//fmt.Printf("SubmitCardBBBBBBBBB:%+v /n", thirteen)
	err = cachethirteen.UpdateGameUser(thirteen, uid, playerNow)
	if err != nil {
		log.Err("thirteen set session failed, %v", err)
		return 0, err
	}
	//fmt.Printf("SubmitCardCCCCCCCCC:%+v /n", thirteen)
	return thirteen.RoomID, nil //

}

func Surrender() {}

func GetGameResult() {

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
		enumr.RoomStatusAllReady)
	if err != nil {
		return nil, err
	}
	thirteens = list
	return thirteens, nil
}
