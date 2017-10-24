package niuniu

import (
	"encoding/json"
	"fmt"
	"math/rand"
	dbbill "playcards/model/bill/db"
	enumbill "playcards/model/bill/enum"
	mdbill "playcards/model/bill/mod"
	cacheniu "playcards/model/niuniu/cache"
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
	"strconv"
	"time"

	"github.com/jinzhu/gorm"
	lua "github.com/yuin/gopher-lua"
)

func CreateNiuniu() []*mdniu.Niuniu {
	rooms, err := GetRoomsByStatusAndGameType()
	if err != nil {
		log.Err("get niuniu rooms by status and game type err :%v", err)
		return nil
	}
	if len(rooms) == 0 {
		return nil
	}
	var newGames []*mdniu.Niuniu
	var userResults []*mdr.GameUserResult
	for _, room := range rooms {
		l := lua.NewState()
		defer l.Close()
		if err := l.DoFile("lua/niuniulua/Logic.lua"); err != nil {
			//fmt.Printf("niuniu logic do file %+v", err)
			log.Err("niuniu logic do file %v", err)
		}
		if err := l.DoString("return Logic:new()"); err != nil {
			//fmt.Printf("niuniu logic do string %+v", err)
			log.Err("niuniu logic do string %v", err)
		}
		//logic := l.Get(1)
		l.Pop(1)

		var roomparam *mdniu.NiuniuRoomParam

		if err := json.Unmarshal([]byte(room.GameParam), &roomparam); err != nil {
			//fmt.Printf("niuniu unmarshal room param failed, %+v", err)
			log.Err("niuniu unmarshal room param failed, %v", err)
			continue
		}
		var status int32
		status = enumniu.GameStatusInit
		var bankerID int32
		// fmt.Printf("AAAA Change Banker:%d|%d|%d\n", room.RoundNow,
		// 	roomparam.BankerType, roomparam.PreBankerID)
		if roomparam.BankerType == enumniu.BankerNoNiu {
			if room.RoundNow > 1 && roomparam.PreBankerID > 0 {
				status = enumniu.GameStatusGetBanker
				bankerID = roomparam.PreBankerID
			}
		} else if roomparam.BankerType == enumniu.BankerTurns {
			status = enumniu.GameStatusGetBanker
			bankerID = SetNextPlayerBanker(room, roomparam.PreBankerID)
		} else if roomparam.BankerType == enumniu.BankerSeeCard {
			bankerID = 0
		} else if roomparam.BankerType == enumniu.BankerDefault {
			status = enumniu.GameStatusGetBanker
			bankerID = room.Users[0].UserID
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
				fmt.Printf("niuniu logic do string %+v", err)
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
							//value, _ := strconv.ParseInt(v.String(), 10, 32)
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
						fmt.Printf("niuniu cardsMap value err %+v",
							value)
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
				fmt.Printf("niuniu cardsMap err %+v", cardsMap)
				log.Err("niuniu cardsMap err %v", cardsMap)
			}
		}

		roomResult := &mdniu.NiuniuRoomResult{
			RoomID: room.RoomID,
			List:   niuUsers,
		}
		now := gorm.NowFunc()
		niuniu := &mdniu.Niuniu{
			RoomID:     room.RoomID,
			Status:     status, //enumniu.GameStatusInit,
			Index:      room.RoundNow,
			BankerType: roomparam.BankerType,
			Result:     roomResult,
			BankerID:   bankerID,
			OpDateAt:   &now,
		}

		f := func(tx *gorm.DB) error {
			err = dbniu.CreateNiuniu(tx, niuniu)
			if err != nil {
				return err
			}

			room.Status = enumr.RoomStatusStarted
			err = cacher.UpdateRoom(room)
			if err != nil {
				fmt.Printf("room update set session failed, %+v", err)
				log.Err("room update set session failed, %v", err)
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
			//fmt.Printf("niuniu create failed,%+v | %+v", niuniu, err)
			log.Err("niuniu create failed,%v | %v", niuniu, err)
			continue
		}
		// now := gorm.NowFunc()
		// niuniu.OpdatedAt = &now
		newGames = append(newGames, niuniu)
		cacheniu.SetGame(niuniu, room.Password)
		if err != nil {
			//fmt.Printf("niuniu create set redis failed,%+v | %+v", room,err)
			log.Err("niuniu create set redis failed,%v | %v", room,
				err)
			continue
		}
		err = cacher.SetRoom(room)
		//fmt.Printf("GameUserResult : %+v\n", room.UserResults)
		if err != nil {
			//fmt.Printf("room create set redis failed,%+v | %+v", room, err)
			log.Err("room create set redis failed,%v | %v", room, err)
			continue
		}
		// for _, user := range room.Users {
		// 	cacheniu.SetGame(room.RoomID, user.UserID)
		// }
	}
	return newGames
}

func UpdateGame() []*mdniu.Niuniu {
	niunius, err := dbniu.GetNiuniuAline(db.DB())
	if err != nil {
		log.Err("get niuniu rooms by status and game type err :%v", err)
		return nil
	}
	if len(niunius) == 0 {
		return nil
	}
	var updateGames []*mdniu.Niuniu
	for _, niuniu := range niunius {
		//fmt.Printf("AAAA Game Status Init:%d", niuniu.Status)
		//isUpdate := false
		if niuniu.Status == enumniu.GameStatusInit {
			sub := time.Now().Sub(*niuniu.OpDateAt)
			//fmt.Printf("AAAA Game Status Init time:%f", sub.Seconds())
			niuniu.BroStatus = enumniu.GameStatusCountDown
			if sub.Seconds() > enumniu.GetBankerTime {
				AutoSetBankerScore(niuniu)
				niuniu.Status = enumniu.GameStatusGetBanker
				UpdateNiuniu(niuniu, true)
			}
		} else if niuniu.Status == enumniu.GameStatusGetBanker {
			if niuniu.BankerID == 0 {
				niuniu.BankerID = ChooseBanker(niuniu)
			}
			var getBankers []*mdniu.GetBanker
			//niuniu.BankerID = bankerID
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
			UpdateNiuniu(niuniu, true)

		} else if niuniu.Status == enumniu.GameStatusSetBet {
			sub := time.Now().Sub(*niuniu.OpDateAt)
			niuniu.BroStatus = enumniu.GameStatusCountDown
			if sub.Seconds() > enumniu.SetBetTime {
				AutoSetBetScore(niuniu)
				niuniu.Status = enumniu.GameStatusAllSetBet
				UpdateNiuniu(niuniu, true)
			}
		} else if niuniu.Status == enumniu.GameStatusAllSetBet {
			niuniu.Status = enumniu.GameStatusSubmitCard
			now := gorm.NowFunc()
			dd, _ := time.ParseDuration("1s")
			now = now.Add(dd)
			niuniu.OpDateAt = &now
			niuniu.BroStatus = enumniu.GameStatusAllSetBet
			UpdateNiuniu(niuniu, true)
		} else if niuniu.Status == enumniu.GameStatusSubmitCard {
			niuniu.BroStatus = enumniu.GameStatusCountDown
			sub := time.Now().Sub(*niuniu.OpDateAt)
			if sub.Seconds() > enumniu.SubmitCardTime {
				niuniu.Status = enumniu.GameStatusStarted
				niuniu.BroStatus = enumniu.GameStatusStarted
				UpdateNiuniu(niuniu, true)
			}
			//niuniu.BroStatus = enumniu.GameStatusDone
		} else if niuniu.Status == enumniu.GameStatusStarted {
			//fmt.Printf("FFF update Started :%d", niuniu.Status)
			pwd := cacheniu.GetRoomPaawordRoomID(niuniu.RoomID)
			room, err := cacher.GetRoom(pwd)
			if err != nil {
				print(err)
				log.Err("room get session failed, %v", err)
				continue
			}
			if room == nil {
				log.Err("room get session nil, %v|%d", pwd, niuniu.RoomID)
				continue
			}

			fmt.Printf("Check Room Param:%v|%s \n", room, pwd)

			l := lua.NewState()
			defer l.Close()
			if err := l.DoFile("lua/niuniulua/Logic.lua"); err != nil {
				log.Err("niuniu clean logic do file %v", err)
				continue
			}

			if err := l.DoString("return Logic:new()"); err != nil {
				log.Err("niuniu get result do string %v", err)
			}
			//fmt.Printf("AAAAAAA:%v", niuniu.GameResults)
			if err := l.DoString(fmt.
				Sprintf("return Logic:CalculateRes('%s','%s')",
					niuniu.GameResults, room.GameParam)); err != nil {
				log.Err("niuniu get result do string %v", err)
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
				m := InitNiuniuTypeMap() //make(map[string]int32)
				for _, userResult := range room.UserResults {
					if userResult.UserID == result.UserID {
						if len(userResult.GameCardCount) > 0 {
							if err := json.Unmarshal([]byte(userResult.GameCardCount), &m); err != nil {
								log.Err("lua str do struct %v", err)
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
						} else if ts == 0 {
							userResult.Lost += 1
						}
					}

				}
				if result.Info.Role == enumniu.Banker {
					if niuniu.BankerType == enumniu.BankerNoNiu{
						if result.Cards.CardType > "0" {
							roomparam.PreBankerID = result.UserID

						} else {
							roomparam.PreBankerID = 0

						}
						data, _ := json.Marshal(&roomparam)
						room.GameParam = string(data)
					}

				}
			}
			niuniu.Status = enumniu.GameStatusDone
			niuniu.BroStatus = enumniu.GameStatusDone
			room.Status = enumr.RoomStatusReInit
			//updateGames = append(updateGames, niuniu)
			f := func(tx *gorm.DB) error {
				niuniu, err = dbniu.UpdateNiuniu(tx, niuniu)
				if err != nil {
					return err
				}
				room, err = dbr.UpdateRoom(tx, room)
				if err != nil {
					return err
				}
				return nil
			}
			err = db.Transaction(f)
			if err != nil {
				print(err)
				log.Err("niuniu update failed, %v", err)
				continue
			}
			//fmt.Printf("GGG :%s", room)
			err = cacheniu.DeleteGame(niuniu.RoomID)
			if err != nil {
				log.Err("niuniu set session failed, %v", err)
				continue
			}
			err = cacher.SetRoom(room)
			if err != nil {
				log.Err("room update room redis failed,%v | %v",
					room, err)
				continue
			}
			//UpdateNiuniu(niuniu, false)
		}

		updateGames = append(updateGames, niuniu)
	}

	return updateGames
}

func UpdateNiuniu(niu *mdniu.Niuniu, cache bool) error {
	if cache {
		err := cacheniu.UpdateGame(niu)
		if err != nil {
			log.Err("niuniu set session failed, %v", err)
			return nil
		}
	}

	f := func(tx *gorm.DB) error {
		_, err := dbniu.UpdateNiuniu(tx, niu)
		if err != nil {
			return err
		}
		return nil
	}
	err := db.Transaction(f)
	if err != nil {
		print(err)
		log.Err("niuniu update failed, %v", err)
		return nil
	}
	return nil
}

func GetRoomsByStatusAndGameType() ([]*mdr.Room, error) {
	var (
		rooms []*mdr.Room
	)
	list, err := dbr.GetRoomsByStatusAndGameType(db.DB(),
		enumr.RoomStatusAllReady, enumniu.GameID)
	if err != nil {
		return nil, err
	}
	rooms = list
	return rooms, nil
}

func GetBanker(uid int32, key int32) (int32, error) {

	//value, err := tools.Int2String(str)
	var value int32
	fmt.Printf("User Get Banker%d", key)
	if v, ok := enumniu.BankerScoreMap[key]; !ok {
		return 0, errorsniu.ErrParam
	} else {
		value = v
	}

	//value := enumniu.BankerScoreMap[key]

	// if value < 0 || value > 4 {
	// 	return 0, errorsniu.ErrParam
	// }

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

	niu, err := cacheniu.GetGame(room.RoomID)

	if niu.Status != enumniu.GameStatusInit {
		return 0, errorsniu.ErrBankerDone
	}

	allReady, userResult := GetUserAndAllOtherStatusReady(niu, uid,
		enumniu.GetBankerStatus)

	if userResult == nil {
		return 0, errorsniu.ErrUserNotInGame
	}

	if userResult.Status > enumniu.UserStatusInit {
		return 0, errorsniu.ErrAlreadyGetBanker
	}

	userResult.Status = enumniu.UserStatusGetBanker
	userResult.Info.BankerScore = value
	if allReady {
		niu.Status = enumniu.GameStatusGetBanker
	}

	f := func(tx *gorm.DB) error {
		niu, err = dbniu.UpdateNiuniu(tx, niu)
		if err != nil {
			return err
		}
		return nil
	}
	err = db.Transaction(f)
	if err != nil {
		return 0, err
	}

	err = cacheniu.UpdateGame(niu)
	if err != nil {
		log.Err("niuniu set session failed, %v", err)
		return 0, err
	}

	return niu.RoomID, nil //

}

func SetBet(uid int32, key int32) (int32, error) {
	var value int32
	if v, ok := enumniu.BetScoreMap[key]; !ok {
		return 0, errorsniu.ErrParam
	} else {
		value = v
	}

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
	niu, err := cacheniu.GetGame(room.RoomID)

	if niu.Status != enumniu.GameStatusSetBet {
		return 0, errorsniu.ErrBetDone
	}

	allReady, userResult := GetUserAndAllOtherStatusReady(niu, uid,
		enumniu.GetBetStatus)

	if userResult == nil {
		return 0, errorsniu.ErrUserNotInGame
	}

	if userResult.Info.Role == enumniu.Banker {
		return 0, errorsniu.ErrBankerNoBet
	}

	if userResult.Status > enumniu.UserStatusGetBanker {
		return 0, errorsniu.ErrAlreadySetBet
	}

	userResult.Status = enumniu.UserStatusSetBet
	userResult.Info.BetScore = value
	if allReady {
		niu.Status = enumniu.GameStatusAllSetBet
	}

	f := func(tx *gorm.DB) error {
		niu, err = dbniu.UpdateNiuniu(tx, niu)
		if err != nil {
			return err
		}

		return nil
	}
	err = db.Transaction(f)
	if err != nil {
		return 0, err
	}

	err = cacheniu.UpdateGame(niu)
	if err != nil {
		log.Err("niuniu set session failed, %v", err)
		return 0, err
	}

	return niu.RoomID, nil //
}

func SubmitCard(uid int32) (int32, error) {
	pwd := cacher.GetRoomPasswordByUserID(uid)
	if len(pwd) == 0 {
		return 0, errors.ErrUserNotInRoom
	}
	room, err := cacher.GetRoom(pwd)
	if err != nil {
		return 0, err
	}

	if room.Status > enumr.RoomStatusStarted {
		if room.Giveup == enumr.WaitGiveUp {
			return 0, errors.ErrInGiveUp
		}
		return 0, errors.ErrGameIsDone
	}

	niu, err := cacheniu.GetGame(room.RoomID)
	if niu == nil {
		return 0, errorsniu.ErrGameNoFind
	}
	if niu.Status != enumniu.GameStatusSubmitCard {
		return 0, errorsniu.ErrSubmitCardDone
	}

	allReady, userResult := GetUserAndAllOtherStatusReady(niu, uid,
		enumniu.GetSubmitCardStatus)

	if userResult == nil {
		return 0, errorsniu.ErrUserNotInGame
	}

	if userResult.Status > enumniu.UserStatusSetBet {
		return 0, errorsniu.ErrAlreadySetBet
	}

	userResult.Status = enumniu.UserStatusSubmitCard

	if allReady {
		niu.Status = enumniu.GameStatusStarted
	}

	f := func(tx *gorm.DB) error {
		niu, err = dbniu.UpdateNiuniu(tx, niu)
		if err != nil {
			return err
		}

		return nil
	}
	err = db.Transaction(f)
	if err != nil {
		return 0, err
	}

	err = cacheniu.UpdateGame(niu)
	if err != nil {
		log.Err("niuniu set session failed, %v", err)
		return 0, err
	}

	return niu.RoomID, nil //

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
	//var ids []int32
	for _, userResult := range niu.Result.List {
		if userResult.Info == nil {
			userResult.Info = &mdniu.BankerAndBet{
				BankerScore: 0,
				BetScore:    0,
				Role:        enumniu.Player,
			}
			userResult.Status = enumniu.UserStatusGetBanker
			//ids = append(ids, userResult.UserID)
		}
	}
	//return ids
}

func AutoSetBetScore(niu *mdniu.Niuniu) {
	//var ids []int32
	for _, userResult := range niu.Result.List {
		if userResult.Info.BetScore == 0 {
			userResult.Info.BetScore = enumniu.MinSetBet
			userResult.Status = enumniu.UserStatusSetBet
			//ids = append(ids, userResult.UserID)
		}
	}
	//return ids
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
	//fmt.Printf("Choose Banker:%v", m)
	ids := m[maxScore]
	if len(ids) > 1 {
		r := rand.New(rand.NewSource(time.Now().UnixNano()))
		id := int32(ids[r.Intn(len(ids))])
		//fmt.Printf("Banker is :%v", id)
		return id
	} else {
		id := ids[0]
		//fmt.Printf("Banker is :%v", id)
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

func CleanGiveUpGame() error {
	var gids []int32
	rids, err := dbr.GetGiveUpRoomIDByGameType(db.DB(), enumniu.GameID)
	if err != nil {
		log.Err("get niuniu give up room err:%v", err)
	}

	for _, rid := range rids {
		game, err := cacheniu.GetGame(rid)
		if err != nil {
			log.Err("get niuniu give up room err:%d|%v", rid, err)
			continue
		}
		if game != nil {
			gids = append(gids, game.GameID)
			err = cacheniu.DeleteGame(rid)
			if err != nil {
				log.Err(" delete niuniu set session failed, %v",
					err)
				continue
			}
		}
		fmt.Printf("CleanGiveUpGame:%+v", game)

	}
	if len(gids) > 0 {
		f := func(tx *gorm.DB) error {
			err = dbniu.GiveUpGameUpdate(tx, gids)
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

func InitNiuniuTypeMap() map[string]int32 {
	m := make(map[string]int32)
	for _, value := range enumniu.NiuniuCardType {
		m[value] = 0
	}
	return m
}
