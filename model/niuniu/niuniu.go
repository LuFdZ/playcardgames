package niuniu

import (
	"encoding/json"
	"fmt"
	"math/rand"
	bill "playcards/model/bill"
	muser "playcards/model/user"
	enumbill "playcards/model/bill/enum"
	mdbill "playcards/model/bill/mod"
	cacheniu "playcards/model/niuniu/cache"
	cacheuser "playcards/model/user/cache"
	dbniu "playcards/model/niuniu/db"
	enumniu "playcards/model/niuniu/enum"
	errorsniu "playcards/model/niuniu/errors"
	mdniu "playcards/model/niuniu/mod"
	cacher "playcards/model/room/cache"
	dbr "playcards/model/room/db"
	enumr "playcards/model/room/enum"
	"playcards/model/room/errors"
	mdr "playcards/model/room/mod"
	pbniu "playcards/proto/niuniu"
	"playcards/utils/db"
	"playcards/utils/log"
	"time"

	"github.com/jinzhu/gorm"
	"github.com/yuin/gopher-lua"
)

func CreateNiuniu() []*mdniu.Niuniu {
	f := func(r *mdr.Room) bool {
		if r.Status == enumr.RoomStatusAllReady && r.GameType == enumniu.GameID {
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
	var newGames []*mdniu.Niuniu

	for _, room := range rooms {
		var userResults []*mdr.GameUserResult
		l := lua.NewState()
		defer l.Close()
		if err := l.DoFile("lua/niuniulua/Logic.lua"); err != nil {
			log.Err("niuniu logic do file %v", err)
		}
		ostimeA := time.Now().UnixNano()
		ostimeB := ostimeA<<32|ostimeA>>32
		if err := l.DoString(fmt.Sprintf("return Logic:new(%d)",ostimeB)); err != nil {
			log.Err("niuniu logic do string %+v", err)
			continue
		}
		l.Pop(1)

		var roomparam *mdniu.NiuniuRoomParam

		if err := json.Unmarshal([]byte(room.GameParam), &roomparam); err != nil {
			log.Err("niuniu unmarshal room param failed, %v", err)
			continue
		}
		var status int32
		status = enumniu.GameStatusInit
		var bankerID int32
		hasNewBanker := true
		if roomparam.BankerType == enumniu.BankerNoNiu {
			if room.RoundNow > 1 && roomparam.PreBankerID > 0 {
				status = enumniu.GameStatusGetBanker
				bankerID = roomparam.PreBankerID
				hasNewBanker = false
			}
		} else if roomparam.BankerType == enumniu.BankerTurns {
			if room.RoundNow > 1 && roomparam.PreBankerID > 0 {
				status = enumniu.GameStatusGetBanker
				bankerID = SetNextPlayerBanker(room, roomparam.PreBankerID)
				hasNewBanker = false
			}
		} else if roomparam.BankerType == enumniu.BankerSeeCard {
			bankerID = 0
		} else if roomparam.BankerType == enumniu.BankerDefault {
			status = enumniu.GameStatusGetBanker
			bankerID = room.Users[0].UserID
			hasNewBanker = false
		}

		var niuUsers []*mdniu.NiuniuUserResult

		for _, user := range room.Users {
			if room.RoundNow == 1 {
				userResult := &mdr.GameUserResult{
					UserID:   user.UserID,
					Nickname: user.Nickname,
					Win:      0,
					Lost:     0,
					Tie:      0,
					Score:    0,
				}
				userResults = append(userResults, userResult)
			}
			if err := l.DoString("return Logic:GetCards()"); err != nil {
				log.Err("niuniu logic do string %v", err)
				continue
			}
			getCards := l.Get(1)
			l.Pop(1)
			var cardList []string
			ctype := "0"
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
						log.Err("niuniu cardsMap value err %v",
							value)
					}
				})
				userCard := &mdniu.UserCard{
					CardList: cardList,
					CardType: ctype,
				}

				userRole := enumniu.Player
				if user.UserID == bankerID {
					userRole = enumniu.Banker
				}
				userInfo := &mdniu.BankerAndBet{
					BankerScore: 0,
					BetScore:    0,
					Role:        int32(userRole),
				}

				niuUser := &mdniu.NiuniuUserResult{
					UserID: user.UserID,
					Status: enumniu.UserStatusInit,
					Cards:  userCard,
					Info:   userInfo,
				}
				niuUsers = append(niuUsers, niuUser)
			} else {
				log.Err("niuniu cardsMap err %v", cardsMap)
			}
		}
		if room.RoundNow == 1 {
			room.UserResults = userResults
		}

		roomResult := &mdniu.NiuniuRoomResult{
			RoomID: room.RoomID,
			List:   niuUsers,
		}
		now := gorm.NowFunc()
		niuniu := &mdniu.Niuniu{
			RoomID:        room.RoomID,
			Status:        status, //enumniu.GameStatusInit,
			Index:         room.RoundNow,
			BankerType:    roomparam.BankerType,
			Result:        roomResult,
			BankerID:      bankerID,
			OpDateAt:      &now,
			HasNewBanker:  hasNewBanker,
			RefreshDateAt: &now,
			Ids:           room.Ids,
		}

		f := func(tx *gorm.DB) error {
			room.Status = enumr.RoomStatusStarted
			_, err := dbr.UpdateRoom(tx, room)
			if err != nil {
				log.Err("Create niuniu room db err:%v|%v", room, err)
				return err
			}
			err = dbniu.CreateNiuniu(tx, niuniu)
			if err != nil {
				log.Err("Create niuniu db err:%v|%v", niuniu, err)
				return err
			}
			err = cacher.UpdateRoom(room)
			if err != nil {
				log.Err("niuniu update room set redis failed,%v | %v", room,
					err)
				return err
			}

			err = cacheniu.SetGame(niuniu, room.Password)
			if err != nil {
				log.Err("niuniu create set redis failed,%v | %v", niuniu,
					err)
				return err
			}

			if room.RoundNow == 1 {
				billForeignKey := fmt.Sprintf("%d:%d:%d", room.GameType, room.RoomID, niuniu.GameID)
				diamond := -int64(room.MaxNumber * room.RoundNumber * enumr.NiuniuGameCost)
				_,u :=cacheuser.GetUserByID(room.PayerID)
				err = muser.SetUserLockBalance(room.PayerID,enumbill.TypeDiamond,-diamond,room.RoomID)
				if err != nil {
					log.Err("room create set lock balance redis failed,%v | %v\n", room, err)
					return err
				}
				_,err := bill.GainBalanceCondition(room.PayerID,u.Channel,u.Version,u.MobileOs,billForeignKey,
					&mdbill.Balance{0, 0, diamond},enumbill.JournalTypeNiuniu)
				if err != nil {
					return err
				}

				for _, user := range room.Users {
					pr := &mdr.PlayerRoom{
						UserID:    user.UserID,
						RoomID:    room.RoomID,
						GameType:  room.GameType,
						PlayTimes: 0,
					}
					err := dbr.CreatePlayerRoom(tx, pr)
					if err != nil {
						log.Err("niuniu create player room err:%v|%v\n", user, err)
						continue
					}
				}
			}
			return nil
		}
		//go db.Transaction(f)

		err := db.Transaction(f)
		if err != nil {
			log.Err("niuniu create failed,%v | %v", niuniu, err)
			continue
		}
		newGames = append(newGames, niuniu)
	}
	return newGames
}

func UpdateGame() []*mdniu.Niuniu {
	f := func(r *mdniu.Niuniu) bool {
		if r.Status < enumniu.GameStatusDone {
			return true
		}
		return false
	}
	niunius := cacheniu.GetAllNiu(f)
	if len(niunius) == 0 {
		return nil
	}
	var updateGames []*mdniu.Niuniu
	for _, niuniu := range niunius {
		if niuniu.Status == enumniu.GameStatusInit {
			sub := time.Now().Sub(*niuniu.OpDateAt)
			niuniu.BroStatus = enumniu.GameStatusCountDown
			if sub.Seconds() > enumniu.GetBankerTime {
				AutoSetBankerScore(niuniu)
				niuniu.Status = enumniu.GameStatusGetBanker
				UpdateNiuniu(niuniu)
			}
		} else if niuniu.Status == enumniu.GameStatusGetBanker {
			if niuniu.BankerID == 0 {
				niuniu.BankerID = ChooseBanker(niuniu)
			}
			var getBankers []*mdniu.GetBanker
			for _, userResult := range niuniu.Result.List {
				gb := &mdniu.GetBanker{
					UserID: userResult.UserID,
					Key:    enumniu.ToBankerScoreMap[userResult.Info.BankerScore],
				}
				getBankers = append(getBankers, gb)
				if userResult.UserID == niuniu.BankerID {
					userResult.Info.Role = enumniu.Banker
					userResult.Status = enumniu.
						UserStatusSetBet
				}
			}
			niuniu.GetBankerList = getBankers
			niuniu.Status = enumniu.GameStatusSetBet
			now := gorm.NowFunc()
			dd, _ := time.ParseDuration("1s")
			now = now.Add(dd)
			niuniu.OpDateAt = &now
			niuniu.BroStatus = enumniu.GameStatusGetBanker
			UpdateNiuniu(niuniu)

		} else if niuniu.Status == enumniu.GameStatusSetBet {
			sub := time.Now().Sub(*niuniu.OpDateAt)
			niuniu.BroStatus = enumniu.GameStatusCountDown
			if sub.Seconds() > enumniu.SetBetTime {
				AutoSetBetScore(niuniu)
				niuniu.Status = enumniu.GameStatusAllSetBet
				UpdateNiuniu(niuniu)
			}
		} else if niuniu.Status == enumniu.GameStatusAllSetBet {
			niuniu.Status = enumniu.GameStatusSubmitCard
			now := gorm.NowFunc()
			dd, _ := time.ParseDuration("1s")
			now = now.Add(dd)
			niuniu.OpDateAt = &now
			niuniu.BroStatus = enumniu.GameStatusAllSetBet
			UpdateNiuniu(niuniu)
		} else if niuniu.Status == enumniu.GameStatusSubmitCard {
			niuniu.BroStatus = enumniu.GameStatusCountDown
			sub := time.Now().Sub(*niuniu.OpDateAt)
			if sub.Seconds() > enumniu.SubmitCardTime {
				niuniu.Status = enumniu.GameStatusStarted
				niuniu.BroStatus = enumniu.GameStatusStarted
				UpdateNiuniu(niuniu)
			}
			//niuniu.BroStatus = enumniu.GameStatusDone
		} else if niuniu.Status == enumniu.GameStatusStarted {
			pwd := cacheniu.GetRoomPaawordRoomID(niuniu.RoomID)
			room, err := cacher.GetRoom(pwd)
			if err != nil {
				//print(err)
				log.Err("niuniu room get session failed, roomid:%d,pwd:%s,err:%v",niuniu.RoomID,pwd, err)
				continue
			}
			if room == nil {
				log.Err("niuniu room get session nil, %v|%d", pwd, niuniu.RoomID)
				continue
			}

			//fmt.Printf("Check Room Param:%v|%s \n", room, pwd)

			l := lua.NewState()
			defer l.Close()
			if err := l.DoFile("lua/niuniulua/Logic.lua"); err != nil {
				log.Err("niuniu clean logic do file %v\n", err)
				continue
			}

			if err := l.DoString(fmt.Sprintf("return Logic:new(%d)",0)); err != nil {
				log.Err("niuniu logic do string %+v", err)
				continue
			}
			niuniu.MarshalNiuniuRoomResult()
			//room.MarshalGameUserResult()
			if err := l.DoString(fmt.
				Sprintf("return Logic:CalculateRes('%s','%s')",
					niuniu.GameResults, room.GameParam)); err != nil {
				log.Err("niuniu return logic calculateres %v|\n%v|\n%v\n",
					niuniu.GameResults, room.GameParam, err)
				continue
			}

			luaResult := l.Get(-1)
			var results *mdniu.NiuniuRoomResult
			//fmt.Printf("niuniu lua result %v: \n", luaResult)
			if err := json.Unmarshal([]byte(luaResult.String()),
				&results); err != nil {
				log.Err("niuniu lua str do struct %v", err)
			}
			niuniu.Result = results
			var roomparam *mdniu.NiuniuRoomParam
			if err := json.Unmarshal([]byte(room.GameParam),
				&roomparam); err != nil {
				log.Err("niuniu unmarshal room param failed, %v",
					err)
				continue
			}
			for _, result := range niuniu.Result.List {
				m := InitNiuniuTypeMap()
				for _, userResult := range room.UserResults {
					if userResult.UserID == result.UserID {
						if len(userResult.GameCardCount) > 0 {
							if err := json.Unmarshal([]byte(userResult.GameCardCount), &m); err != nil {
								log.Err("niuniu lua str do struct %v", err)
							}
						}
						if _, ok := m[result.Cards.CardType]; ok {
							m[result.Cards.CardType]++
						}
						r, _ := json.Marshal(m)
						userResult.GameCardCount = string(r)
						ts := result.Score
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
				if result.Info.Role == enumniu.Banker {
					roomparam.PreBankerID = result.UserID
					if niuniu.BankerType == enumniu.BankerNoNiu {
						if result.Cards.CardType == "0" {
							roomparam.PreBankerID = 0
						}
					}
				}
			}
			data, _ := json.Marshal(&roomparam)
			room.GameParam = string(data)
			niuniu.Status = enumniu.GameStatusDone
			niuniu.BroStatus = enumniu.GameStatusDone
			room.Status = enumr.RoomStatusReInit
			//updateGames = append(updateGames, niuniu)
			f := func(tx *gorm.DB) error {
				niuniu, err = dbniu.UpdateNiuniu(tx, niuniu)
				if err != nil {
					log.Err("niuniu update db failed, %v|%v", niuniu, err)
					return err
				}
				room, err = dbr.UpdateRoom(tx, room)
				if err != nil {
					log.Err("niuniu update room db failed, %v|%v", niuniu, err)
					return err
				}
				return nil
			}
			//go db.Transaction(f)
			err = db.Transaction(f)
			if err != nil {
				//print(err)
				log.Err("niuniu update failed, %v", err)
				continue
			}
			err = cacheniu.DeleteGame(niuniu.RoomID)
			if err != nil {
				log.Err("niuniu room del session failed, roomid:%d,pwd:%s,err:%v",niuniu.RoomID,pwd, err)
				continue
			}

			err = cacher.UpdateRoom(room)
			if err != nil {
				log.Err("niuniu room update room redis failed,%v | %v",
					room, err)
				continue
			}
			//UpdateNiuniu(niuniu, false)
		}
		updateGames = append(updateGames, niuniu)
	}

	return updateGames
}

func UpdateNiuniu(niu *mdniu.Niuniu) error {
	err := cacheniu.UpdateGame(niu)
	if err != nil {
		log.Err("niuniu set session failed, %v", err)
		return nil
	}
	return nil
}

//func GetRoomsByStatusAndGameType() ([]*mdr.Room, error) {
//	var (
//		rooms []*mdr.Room
//	)
//	list, err := dbr.GetRoomsByStatusAndGameType(db.DB(),
//		enumr.RoomStatusAllReady, enumniu.GameID)
//	if err != nil {
//		return nil, err
//	}
//	rooms = list
//	return rooms, nil
//}

func GetBanker(uid int32, key int32,room *mdr.Room) error {

	//value, err := tools.Int2String(str)
	var value int32
	if v, ok := enumniu.BankerScoreMap[key]; !ok {
		return errorsniu.ErrParam
	} else {
		value = v
	}

	//value := enumniu.BankerScoreMap[key]

	// if value < 0 || value > 4 {
	// 	return 0, errorsniu.ErrParam
	// }

	if room.Status > enumr.RoomStatusStarted {
		return errors.ErrGameIsDone
	}

	if room.Giveup == enumr.WaitGiveUp {
		return errors.ErrInGiveUp
	}

	niu, err := cacheniu.GetGame(room.RoomID)

	if niu.Status != enumniu.GameStatusInit {
		return errorsniu.ErrBankerDone
	}

	allReady, userResult := GetUserAndAllOtherStatusReady(niu, uid,
		enumniu.GetBankerStatus)

	if userResult == nil {
		return errorsniu.ErrUserNotInGame
	}

	if userResult.Status > enumniu.UserStatusInit {
		return errorsniu.ErrAlreadyGetBanker
	}

	userResult.Status = enumniu.UserStatusGetBanker
	userResult.Info.BankerScore = value
	if allReady {
		niu.Status = enumniu.GameStatusGetBanker
	}

	err = cacheniu.UpdateGame(niu)
	if err != nil {
		log.Err("niuniu set session failed, %v", err)
		return err
	}

	//读写分离
	//f := func(tx *gorm.DB) error {
	//	niu, err = dbniu.UpdateNiuniu(tx, niu)
	//	if err != nil {
	//		return err
	//	}
	//	return nil
	//}
	//err = db.Transaction(f)
	//if err != nil {
	//	return 0, err
	//}
	//读写分离

	return nil //

}

func SetBet(uid int32, key int32,room *mdr.Room) ([]int32, error) {
	var value int32
	if v, ok := enumniu.BetScoreMap[key]; !ok {
		return nil, errorsniu.ErrParam
	} else {
		value = v
	}


	if room.Status > enumr.RoomStatusStarted {

		return nil, errors.ErrGameIsDone
	}
	if room.Giveup == enumr.WaitGiveUp {
		return nil, errors.ErrInGiveUp
	}
	niu, err := cacheniu.GetGame(room.RoomID)

	if niu.Status != enumniu.GameStatusSetBet {
		return nil, errorsniu.ErrBetDone
	}

	allReady, userResult := GetUserAndAllOtherStatusReady(niu, uid,
		enumniu.GetBetStatus)

	if userResult == nil {
		return nil, errorsniu.ErrUserNotInGame
	}

	if userResult.Info.Role == enumniu.Banker {
		return nil, errorsniu.ErrBankerNoBet
	}

	if userResult.Status > enumniu.UserStatusGetBanker {
		return nil, errorsniu.ErrAlreadySetBet
	}

	userResult.Status = enumniu.UserStatusSetBet
	userResult.Info.BetScore = value
	if allReady {
		niu.Status = enumniu.GameStatusAllSetBet
	}

	err = cacheniu.UpdateGame(niu)
	if err != nil {
		log.Err("niuniu set session failed, %v", err)
		return nil, err
	}

	//f := func(tx *gorm.DB) error {
	//	niu, err = dbniu.UpdateNiuniu(tx, niu)
	//	if err != nil {
	//		return err
	//	}
	//
	//	return nil
	//}
	//err = db.Transaction(f)
	//if err != nil {
	//	return 0, err
	//}

	return niu.Ids, nil //
}

func SubmitCard(uid int32,room *mdr.Room) ([]int32, error) {
	if room.Status > enumr.RoomStatusStarted {
		if room.Giveup == enumr.WaitGiveUp {
			return nil, errors.ErrInGiveUp
		}
		return nil, errors.ErrGameIsDone
	}

	niu, err := cacheniu.GetGame(room.RoomID)
	if niu == nil {
		return nil, errorsniu.ErrGameNoFind
	}
	if niu.Status != enumniu.GameStatusSubmitCard {
		return nil, errorsniu.ErrSubmitCardDone
	}

	allReady, userResult := GetUserAndAllOtherStatusReady(niu, uid,
		enumniu.GetSubmitCardStatus)

	if userResult == nil {
		return nil, errorsniu.ErrUserNotInGame
	}

	if userResult.Status > enumniu.UserStatusSetBet {
		return nil, errorsniu.ErrAlreadySetBet
	}

	userResult.Status = enumniu.UserStatusSubmitCard

	if allReady {
		niu.Status = enumniu.GameStatusStarted
	}

	err = cacheniu.UpdateGame(niu)
	if err != nil {
		log.Err("niuniu set session failed, %v", err)
		return nil, err
	}
	return niu.Ids, nil //
}

func GetUserAndAllOtherStatusReady(n *mdniu.Niuniu, uid int32,
	getType int32) (bool, *mdniu.NiuniuUserResult) {
	var userResult *mdniu.NiuniuUserResult
	allReady := true
	for _, user := range n.Result.List {
		if user.UserID == uid {
			userResult = user
			if getType == enumniu.GetBetStatus &&
				userResult.Info.Role == enumniu.Banker {
				return false, userResult
			}
		} else if getType == enumniu.GetBankerStatus {
			if user.Status != enumniu.UserStatusGetBanker {
				allReady = false
			}
		} else if getType == enumniu.GetBetStatus {
			if user.Info.Role == enumniu.Banker {
				user.Status = enumniu.UserStatusSetBet
			}
			if user.Status != enumniu.UserStatusSetBet {
				allReady = false
			}
		} else if getType == enumniu.GetSubmitCardStatus {
			if user.Status != enumniu.UserStatusSubmitCard {
				allReady = false
			}
		}
	}
	return allReady, userResult
}

func AutoSetBankerScore(niu *mdniu.Niuniu) {
	for _, userResult := range niu.Result.List {
		if userResult.Status == enumniu.UserStatusInit {
			userResult.Info = &mdniu.BankerAndBet{
				BankerScore: 0,
				BetScore:    0,
				Role:        enumniu.Player,
			}
			userResult.Status = enumniu.UserStatusGetBanker
		}
	}
}

func AutoSetBetScore(niu *mdniu.Niuniu) {
	for _, userResult := range niu.Result.List {
		if userResult.Info.BetScore == 0 {
			userResult.Info.BetScore = enumniu.MinSetBet
			userResult.Status = enumniu.UserStatusSetBet
		}
	}
}

func ChooseBanker(niu *mdniu.Niuniu) int32 {
	var m map[int32][]int32
	m = make(map[int32][]int32)
	var maxScore int32
	for _, user := range niu.Result.List {
		ids := m[user.Info.BankerScore]
		if ids == nil {
			ids = []int32{}
		}
		ids = append(ids, user.UserID)
		m[user.Info.BankerScore] = ids
		if user.Info.BankerScore > maxScore {
			maxScore = user.Info.BankerScore
		}
	}
	ids := m[maxScore]
	if len(ids) > 1 {
		r := rand.New(rand.NewSource(time.Now().UnixNano()))
		id := int32(ids[r.Intn(len(ids))])
		return id
	} else {
		id := ids[0]
		return id
	}
}

func SetNextPlayerBanker(room *mdr.Room, bankerLast int32) int32 {
	var bankerID int32
	for i := 0; i < len(room.Users); i++ {
		room.Users[i].Ready = enumr.UserUnready
		if room.Users[i].UserID == bankerLast {
			if i == len(room.Users)-1 {
				room.Users[0].Role = enumr.UserRoleMaster
				bankerID = room.Users[0].UserID
			} else {
				room.Users[i+1].Role = enumr.UserRoleMaster
				bankerID = room.Users[i+1].UserID
			}
		}
	}
	return bankerID
}

func GameResultList(rid int32) (*pbniu.GameResultListReply, error) {
	var list []*pbniu.NiuniuRoomResult
	niunius, err := dbniu.GetNiuniuByRoomID(db.DB(), rid)
	if err != nil {
		return nil, err
	}
	for _, niuniu := range niunius {
		result := niuniu.Result
		list = append(list, result.ToProto())
	}
	out := &pbniu.GameResultListReply{
		List: list,
	}
	return out, nil
}

func NiuniuRecovery(rid int32, uid int32) (int32, *mdniu.NiuniuRoomResult,
	error) {
	niu, err := cacheniu.GetGame(rid)
	if err != nil {
		return 0, nil, err
	}
	if niu == nil {
		niu, err = dbniu.GetLastNiuniuByRoomID(db.DB(), rid)
		if err != nil {
			return 0, nil, err
		}
	}
	if niu == nil {
		return 0, nil, errorsniu.ErrGameNotExist
	}
	return niu.Status, niu.Result, nil
}

func CleanGame() error {
	var gids []int32
	rids, err := cacher.GetAllDeleteRoomKey(enumniu.GameID)
	if err != nil {
		log.Err("get niuniu clean room err:%v", err)
		return err
	}
	for _, rid := range rids {
		game, err := cacheniu.GetGame(rid)
		if err != nil {
			log.Err("get niuniu give up room err:%d|%v", rid, err)
			continue
		}
		if game != nil {
			log.Debug("clean niuniu game:%d|%d|%+v\n",game.GameID,game.RoomID,game.Ids)
			gids = append(gids, game.GameID)
			err = cacheniu.DeleteGame(rid)
			if err != nil {
				log.Err(" delete niuniu set session failed, %v",
					err)
				continue
			}
			err = cacher.CleanDeleteRoom(enumniu.GameID,game.RoomID)
			if err != nil {
				log.Err(" delete niuniu delete room session failed,roomid:%d,err: %v",game.RoomID,
					err)
				continue
			}
		}else{
			err = cacher.CleanDeleteRoom(enumniu.GameID,rid)
			if err != nil {
				log.Err(" delete null game niuniu delete room session failed,roomid:%d,err: %v",rid,
					err)
				continue
			}
		}
	}
	if len(gids) > 0 {
		f := func(tx *gorm.DB) error {
			err = dbniu.GiveUpGameUpdate(tx, gids)
			if err != nil {
				return err
			}
			return nil
		}
		go db.Transaction(f)
	}

	return nil
}

func InitNiuniuTypeMap() map[string]int32 {
	m := make(map[string]int32)
	for _, value := range enumniu.NiuniuCardType {
		m[value] = 0
	}
	return m
}
