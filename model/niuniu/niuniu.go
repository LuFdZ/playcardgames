package niuniu

import (
	"encoding/json"
	"fmt"
	"math/rand"
	cacheniu "playcards/model/niuniu/cache"
	dbniu "playcards/model/niuniu/db"
	enumniu "playcards/model/niuniu/enum"
	errorsniu "playcards/model/niuniu/errors"
	mdniu "playcards/model/niuniu/mod"
	cacher "playcards/model/room/cache"
	dbr "playcards/model/room/db"
	//enumr "playcards/model/room/enum"
	errroom "playcards/model/room/errors"
	mdroom "playcards/model/room/mod"
	pbniu "playcards/proto/niuniu"
	//gsync "playcards/utils/sync"
	enumroom "playcards/model/room/enum"
	"playcards/model/room"
	"playcards/utils/db"
	"playcards/utils/log"
	"time"

	"github.com/jinzhu/gorm"
	"github.com/yuin/gopher-lua"
	//"playcards/utils/tools"
	"playcards/utils/tools"
)

//var GoLua *lua.LState
//
//func InitGoLua(gl *lua.LState) {
//	GoLua = gl
//}

func RoomLockKey(pwd string) string {
	return fmt.Sprintf("playcards.room.op.lock:%s", pwd)
}

func CreateNiuniu(goLua *lua.LState) []*mdniu.Niuniu {
	rooms := cacher.GetAllRoomByGameTypeAndStatus(enumroom.NiuniuGameType, enumroom.RoomStatusAllReady)
	if rooms == nil && len(rooms) == 0 {
		return nil
	}
	if len(rooms) == 0 {
		return nil
	}
	var newGames []*mdniu.Niuniu

	for _, mdr := range rooms {
		var roomparam *mdroom.NiuniuRoomParam

		if err := json.Unmarshal([]byte(mdr.GameParam), &roomparam); err != nil {
			log.Err("niuniu unmarshal room param failed, %v", err)
			continue
		}
		var status int32
		status = enumniu.GameStatusInit
		var bankerID int32
		hasNewBanker := true
		if roomparam.BankerType == enumniu.BankerNoNiu {
			if mdr.RoundNow > 1 && roomparam.PreBankerID > 0 {
				status = enumniu.GameStatusGetBanker
				bankerID = roomparam.PreBankerID
				hasNewBanker = false
			}
		} else if roomparam.BankerType == enumniu.BankerTurns {
			if mdr.RoundNow > 1 && roomparam.PreBankerID > 0 {
				status = enumniu.GameStatusGetBanker
				bankerID = SetNextPlayerBanker(mdr, roomparam.PreBankerID)
				hasNewBanker = false
			}
		} else if roomparam.BankerType == enumniu.BankerSeeCard {
			bankerID = 0
		} else if roomparam.BankerType == enumniu.BankerDefault {
			status = enumniu.GameStatusGetBanker
			bankerID = mdr.Users[0].UserID
			hasNewBanker = false
		} else if roomparam.BankerType == enumniu.BankerAll {
			status = enumniu.GameStatusAllSetBet
			bankerID = 0
			hasNewBanker = false
		}

		var userResults []*mdroom.GameUserResult
		var niuUsers []*mdniu.NiuniuUserResult
		userResults, niuUsers, err := InitUserCard(mdr, bankerID, userResults, niuUsers, roomparam.AdvanceOptions[1], goLua)
		//fmt.Printf("CreateNiuniu:%v|%v|%v\n", userResults, niuUsers, err)
		if err != nil {
			continue
		}
		if mdr.RoundNow == 1 {
			mdr.UserResults = userResults
		}
		if mdr.RoundNumber > 1 && roomparam.AdvanceOptions[0] != "0" {
			for _, ru := range mdr.UserResults {
				if ru.LastRole == enumniu.Player && ru.RoundScore > 0 && (mdr.RoundNow-ru.LastPushOnBet) > 1 {
					score := ru.RoundScore + 4*roomparam.BetScore
					scoreLimit := enumniu.PushOnScoreMap[roomparam.AdvanceOptions[0]] * roomparam.BetScore
					if score > scoreLimit {
						score = scoreLimit
					}
					if score < 4 {
						ru.CanPushOn = enumroom.NoPushOnBet
						ru.PushOnBetScore = -1
					} else {
						ru.PushOnBetScore = score
						ru.CanPushOn = enumroom.PushOnBet
						for _, niuUser := range niuUsers {
							if niuUser.UserID == ru.UserID {
								niuUser.PushOnBet = ru.PushOnBetScore
								break
							}
						}
					}
				} else {
					ru.CanPushOn = enumroom.NoPushOnBet
					ru.PushOnBetScore = -1
				}
				log.Debug("DDDDreateNiuniuAdvanceOptions:%d|%d\n", ru.UserID, ru.CanPushOn)
			}

		}

		roomResult := &mdniu.NiuniuRoomResult{
			RoomID: mdr.RoomID,
			List:   niuUsers,
			//RobotIds: mdr.RobotIds,
		}

		now := gorm.NowFunc()
		niuniu := &mdniu.Niuniu{
			RoomID:        mdr.RoomID,
			Status:        status, //enumniu.GameStatusInit,
			Index:         mdr.RoundNow,
			BankerType:    roomparam.BankerType,
			Result:        roomResult,
			BankerID:      bankerID,
			OpDateAt:      &now,
			HasNewBanker:  hasNewBanker,
			RefreshDateAt: &now,
			PassWord:      mdr.Password,
			RoomType:      mdr.RoomType,
			Ids:           mdr.Ids,
			RobotIds:      mdr.RobotIds,
			RobotOpMap:    make(map[int32][]int32),
			RoomParam:     roomparam,
		}
		//for _, user := range mdr.Users {
		//	if mdr.RoomType == enumroom.RoomTypeGold && user.Ready == enumroom.UserUnready {
		//		niuniu.WatchIds = append(niuniu.WatchIds, user.UserID)
		//	}
		//}
		f := func(tx *gorm.DB) error {
			if mdr.RoundNow == 1 {
				//if room.RoomType != enumr.RoomTypeClub && room.Cost != 0 {
				//	err := bill.GainGameBalance(room.PayerID, room.RoomID, enumbill.JournalTypeNiuniu,
				//		enumbill.JournalTypeNiuniuUnFreeze, &mdbill.Balance{Amount: room.Cost, CoinType: room.CostType})
				//	if err != nil {
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
					err := dbr.CreatePlayerRoom(tx, pr)
					if err != nil {
						log.Err("niuniu create player room err:%v|%v\n", user, err)
						continue
					}
				}
			}
			mdr.Status = enumroom.RoomStatusStarted
			_, err := dbr.UpdateRoom(tx, mdr)
			if err != nil {
				log.Err("Create niuniu room db err:%v|%v", mdr, err)
				return err
			}
			err = dbniu.CreateNiuniu(tx, niuniu)
			if err != nil {
				log.Err("Create niuniu db err:%v|%v", niuniu, err)
				return err
			}
			err = cacher.UpdateRoom(mdr)
			if err != nil {
				log.Err("niuniu update room set redis failed,%v | %v", mdr,
					err)
				return err
			}

			err = cacheniu.SetGame(niuniu)
			if err != nil {
				log.Err("niuniu create set redis failed,%v | %v", niuniu,
					err)
				return err
			}

			return nil
		}
		//go db.Transaction(f)

		err = db.Transaction(f)
		if err != nil {
			log.Err("niuniu create failed,%v | %v", niuniu, err)
			continue
		}
		newGames = append(newGames, niuniu)
	}
	return newGames
}

func InitUserCard(mdr *mdroom.Room, bankerID int32, userResults []*mdroom.GameUserResult,
	niuUsers []*mdniu.NiuniuUserResult, noFaceCard string, goLua *lua.LState) ([]*mdroom.GameUserResult, []*mdniu.NiuniuUserResult, error) {

	if err := goLua.DoString(fmt.Sprintf("return G_Reset(%s)", noFaceCard)); err != nil {
		log.Err("niuniu G_Reset %+v", err)
		return nil, nil, errorsniu.ErrGoLua
	}
	for _, user := range mdr.Users {
		if mdr.RoomType == enumroom.RoomTypeGold && user.Ready == enumroom.UserUnready {
			continue
		}
		if mdr.RoundNow == 1 {
			userResult := &mdroom.GameUserResult{
				UserID: user.UserID,
				//Nickname: user.Nickname,
				Win:            0,
				Lost:           0,
				Tie:            0,
				Score:          0,
				LastPushOnBet:  0,
				PushOnBetScore: -1,
				CanPushOn:      enumroom.NoPushOnBet,
			}
			userResults = append(userResults, userResult)
		}
		if err := goLua.DoString("return G_GetCards()"); err != nil {
			log.Err("niuniu G_GetCards err %v", err)
			return nil, nil, errorsniu.ErrGoLua
		}
		getCards := goLua.Get(-1)
		goLua.Pop(1)
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
			if len(cardList) == 0 {
				log.Err("niuniu cardList nil err %v", cardsMap)
				return nil, nil, errorsniu.ErrGoLua
			}
			userCard := &mdniu.UserCard{
				CardList: cardList,
				CardType: ctype,
			}

			userRole := enumniu.Player
			if user.UserID == bankerID {
				userRole = enumniu.Banker
			}
			userInfo := &mdniu.BankerAndBet{
				BankerScore: 1,
				BetScore:    0,
				Role:        int32(userRole),
			}
			t := time.Now()
			niuUser := &mdniu.NiuniuUserResult{
				UserID:    user.UserID,
				Status:    enumniu.UserStatusInit,
				Cards:     userCard,
				Info:      userInfo,
				Type:      user.Type,
				PushOnBet: -1,
				UpdateAt:  &t,
			}
			niuUsers = append(niuUsers, niuUser)
		} else {
			log.Err("niuniu cardsMap err %v", cardsMap)
			return nil, nil, errorsniu.ErrGoLua
		}
	}
	return userResults, niuUsers, nil
}

func UpdateGame(goLua *lua.LState) []*mdniu.Niuniu {
	niunius, err := cacheniu.GetAllNiuniuByStatus(enumniu.GameStatusDone)
	if err != nil {
		log.Err("get all niuniu by status err err:%v", err)
		return nil
	}
	if len(niunius) == 0 {
		return nil
	}

	var updateGames []*mdniu.Niuniu
	for _, niuniu := range niunius {
		if niuniu.Status < enumniu.GameStatusDone && niuniu.RoomType == enumroom.RoomTypeGold && len(niuniu.RobotIds) > 0 {
			niuniu.RobotOpMap = make(map[int32][]int32)
		}
		if niuniu.Status == enumniu.GameStatusInit {
			sub := time.Now().Sub(*niuniu.OpDateAt)
			niuniu.BroStatus = enumniu.GameStatusCountDown
			if sub.Seconds() > enumniu.GetBankerTime {
				AutoSetBankerScore(niuniu)
				niuniu.Status = enumniu.GameStatusGetBanker
				UpdateNiuniu(niuniu)
			} else if niuniu.RoomType == enumroom.RoomTypeGold && len(niuniu.RobotIds) > 0 && niuniu.RobotOpStatus == 0 {
				for _, robotID := range niuniu.RobotIds {
					for _, u := range niuniu.Result.List {
						if u.UserID == robotID && u.Status == enumniu.UserStatusInit {
							robotSubmitTime := float64(tools.GenerateRangeNum(4, 7))
							if sub.Seconds() > robotSubmitTime {
								var robotKey int32 = 1
								niu, err := GetBanker(robotID, robotKey)
								if err != nil {
									log.Err("robot set banker err robotID:%d,err:%v\n", robotID, err)
								}
								niuniu = niu
								niuniu.RobotOpStatus = enumniu.UserStatusGetBanker
								UpdateNiuniu(niuniu)
							} else {
								niuniu.RobotOpStatus = 0
							}
						}
					}
				}
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
					//if userResult.Info.BankerScore == 0{
					//	userResult.Info.BankerScore =1
					//}
					userResult.Status = enumniu.UserStatusSetBet
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
			} else if niuniu.RoomType == enumroom.RoomTypeGold && len(niuniu.RobotIds) > 0 && niuniu.RobotOpStatus == enumniu.UserStatusGetBanker {
				//niuniu.RobotOpStatus = enumniu.UserStatusSetBet
				for _, robotID := range niuniu.RobotIds {
					for _, u := range niuniu.Result.List {
						if u.UserID == robotID && u.Status == enumniu.UserStatusGetBanker && u.Info.Role != enumniu.Banker {
							robotSubmitTime := float64(tools.GenerateRangeNum(4, 7))
							fmt.Printf("AAAAUserStatusGetBanker:%f|%f\n", sub.Seconds(), robotSubmitTime)
							if sub.Seconds() > robotSubmitTime {
								fmt.Printf("BBBBUserStatusGetBanker:%f|%f\n", sub.Seconds(), robotSubmitTime)
								robotKey := tools.GenerateRangeNum(1, 5)
								//u.Info.BetScore = robotKey
								niu, err := SetBet(robotID, robotKey)
								if err != nil {
									log.Err("robot set bet err robotID:%d,err:%v\n", robotID, err)
								}
								niuniu = niu
								niuniu.RobotOpStatus = enumniu.UserStatusSetBet
								niuniu.RobotOpMap[u.UserID] = []int32{enumniu.UserStatusSetBet, robotKey}
								UpdateNiuniu(niuniu)
								//u.Status = enumniu.UserStatusSetBet
							} else {
								niuniu.RobotOpStatus = enumniu.UserStatusGetBanker
							}
						}
					}
					//UpdateNiuniu(niuniu)
				}
			}
		} else if niuniu.Status == enumniu.GameStatusAllSetBet {
			if niuniu.BankerType == enumniu.BankerAll {
				for _, userResult := range niuniu.Result.List {
					if userResult.Status == enumniu.UserStatusInit {
						userResult.Info = &mdniu.BankerAndBet{
							BankerScore: 1,
							BetScore:    niuniu.RoomParam.BetScore,
							Role:        enumniu.Player,
						}
						userResult.Status = enumniu.UserStatusGetBanker
					}
				}
			}

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
			} else if niuniu.RoomType == enumroom.RoomTypeGold && len(niuniu.RobotIds) > 0 && niuniu.RobotOpStatus == enumniu.UserStatusSetBet {
				//niuniu.RobotOpStatus = enumniu.UserStatusSubmitCard
				for _, robotID := range niuniu.RobotIds {
					for _, u := range niuniu.Result.List {
						if u.UserID == robotID && u.Status == enumniu.UserStatusSetBet {
							robotSubmitTime := float64(tools.GenerateRangeNum(4, 7))
							fmt.Printf("AAAAUserStatusSubmitCard:%f|%f\n", sub.Seconds(), robotSubmitTime)
							if sub.Seconds() > robotSubmitTime {
								fmt.Printf("BBBBUserStatusSubmitCard:%f|%f\n", sub.Seconds(), robotSubmitTime)
								niu, err := SubmitCard(robotID)
								if err != nil {
									log.Err("robot submit err robotID:%d,err:%v\n", robotID, err)
								}
								niuniu = niu
								niuniu.RobotOpStatus = enumniu.UserStatusSubmitCard
								//u.Status = enumniu.UserStatusSubmitCard
								niuniu.RobotOpMap[u.UserID] = []int32{enumniu.UserStatusSubmitCard, 0}
								UpdateNiuniu(niuniu)
							} else {
								niuniu.RobotOpStatus = enumniu.UserStatusSetBet
							}
						}
					}
					//UpdateNiuniu(niuniu)
				}
			}
			//niuniu.BroStatus = enumniu.GameStatusDone
		} else if niuniu.Status == enumniu.GameStatusStarted {
			mdr, err := cacher.GetRoom(niuniu.PassWord)
			if err != nil {
				//print(err)
				log.Err("niuniu room get session failed, roomid:%d,pwd:%s,err:%v", niuniu.RoomID, niuniu.PassWord, err)
				err = cacheniu.DeleteGame(niuniu)
				if err != nil {
					log.Err("niuniu room not exist delete  set session failed, %v",
						err)
				}
				log.Err("niuniu room not exist delete  game, %v",
					niuniu)
				continue
			}
			if mdr == nil {
				log.Err("niuniu room get session nil, %v|%d", niuniu.PassWord, niuniu.RoomID)
				continue
			}
			for _, u := range niuniu.Result.List {
				if u.Info.BetScore == 0 {
					u.Info.BetScore = 1
				}
			}
			niuniu.MarshalNiuniuRoomResult()
			//room.MarshalGameUserResult()
			if err := goLua.DoString(fmt.
			Sprintf("return G_CalculateRes('%s','%s')",
				niuniu.GameResults, mdr.GameParam)); err != nil {
				log.Err("niuniu G_CalculateRes %v|\n%v|\n%v\n",
					niuniu.GameResults, mdr.GameParam, err)
				continue
			}

			luaResult := goLua.Get(-1)
			goLua.Pop(1)
			var results *mdniu.NiuniuRoomResult
			if err := json.Unmarshal([]byte(luaResult.String()),
				&results); err != nil {
				log.Err("niuniu lua str do struct %v", err)
			}
			//niuniu.Result = results

			var roomparam *mdroom.NiuniuRoomParam
			if err := json.Unmarshal([]byte(mdr.GameParam),
				&roomparam); err != nil {
				log.Err("niuniu unmarshal room param failed, %v",
					err)
				continue
			}
			var specialCardUids []int32
			for i, result := range niuniu.Result.List {
				result.Score = results.List[i].Score
				result.Cards.CardType = results.List[i].Cards.CardType
				m := InitNiuniuTypeMap()
				for _, userResult := range mdr.UserResults {
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
						//特殊记录 五小牛
						if result.Cards.CardType == "13" {
							specialCardUids = append(specialCardUids, userResult.UserID)
						}
						for _, ru := range mdr.Users {
							if ru.UserID == userResult.UserID {
								ru.ResultAmount = userResult.Score
								break
							}
						}
						userResult.RoundScore = ts
						if niuniu.RoomParam.AdvanceOptions[0] != "0" {
							userResult.LastRole = result.Info.Role
							//if userResult.LastPushOnBet == enumroom.PushOnBet &&
							//	userResult.CanPushOn == enumroom.NoPushOnBet {
							//	userResult.LastPushOnBet = enumroom.NoPushOnBet
							//}
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
			mdr.GameParam = string(data)
			mdr.GameIDNow = niuniu.GameID
			niuniu.Status = enumniu.GameStatusDone
			niuniu.BroStatus = enumniu.GameStatusDone
			mdr.Status = enumroom.RoomStatusReInit

			if mdr.RoomType == enumroom.RoomTypeClub && mdr.SubRoomType == enumroom.SubTypeClubMatch {
				err := room.GetRoomClubCoin(mdr)
				if err != nil {
					log.Err("room club member game balance failed,rid:%d,uid:%d, err:%v", mdr.RoomID, err)
					continue
				}
				for _, ur := range mdr.UserResults {
					for _, ugr := range niuniu.Result.List {
						if ugr.UserID == ugr.UserID {
							ugr.ClubCoinScore = ur.RoundClubCoinScore
							break
						}
					}
				}
			}
			f := func(tx *gorm.DB) error {
				niuniu, err = dbniu.UpdateNiuniu(tx, niuniu)
				if err != nil {
					log.Err("niuniu update db failed, %v|%v", niuniu, err)
					return err
				}
				mdr, err = dbr.UpdateRoom(tx, mdr)
				if err != nil {
					log.Err("niuniu update room db failed, %v|%v", niuniu, err)
					return err
				}
				for _, uid := range specialCardUids {
					plsgr := &mdroom.PlayerSpecialGameRecord{
						GameID:     niuniu.GameID,
						RoomID:     niuniu.RoomID,
						GameType:   mdr.GameType,
						RoomType:   mdr.RoomType,
						Password:   mdr.Password,
						UserID:     uid,
						GameResult: niuniu.GameResults,
					}
					err = dbr.CreateSpecialGame(tx, plsgr)
					if err != nil {
						return err
					}
				}
				return nil
			}
			err = db.Transaction(f)
			if err != nil {
				log.Err("niuniu update failed, %v", err)
				continue
			}
			err = cacheniu.DeleteGame(niuniu)
			if err != nil {
				log.Err("niuniu room del session failed, roomid:%d,pwd:%s,err:%v", niuniu.RoomID, niuniu.PassWord, err)
				continue
			}

			err = cacher.UpdateRoom(mdr)
			if err != nil {
				log.Err("niuniu room update room redis failed,%v | %v",
					mdr, err)
				continue
			}
			//UpdateNiuniu(niuniu, false)
		}
		updateGames = append(updateGames, niuniu)
	}
	return updateGames
}

//func UpdateRobotGame() (map[*mdniu.Niuniu][]int32) {
//	niunius, err := cacheniu.GetRobotNiuniuByStatus(enumniu.GameStatusDone)
//	if err != nil {
//		log.Err("GetAllNiuniuByStatusErr err:%v", err)
//		return nil
//	}
//	//fmt.Printf("UpdateRobotGame:%d\n",len(niunius))
//	if len(niunius) == 0 {
//		return nil
//	}
//	m := make(map[*mdniu.Niuniu][]int32)
//	for _, niuniu := range niunius {
//		for _, niuUser := range niuniu.Result.List {
//			if niuUser.Type != enumr.Robot {
//				continue
//			}
//			subTime := time.Now().Sub(*niuUser.UpdateAt)
//			addTime := float64(tools.GenerateRangeNum(4, 9))
//			//fmt.Printf("UpdateRobotGame:%d|%f\n",niuniu.Status,subTime.Seconds())
//			if niuniu.Status == enumniu.GameStatusInit && niuUser.Status == enumniu.UserStatusInit {
//				if subTime.Seconds() > addTime {
//					f := func() error {
//						err := GetBanker(niuUser.UserID, int32(addTime-4))
//						if err != nil {
//							return nil
//						}
//						return nil
//					}
//					lock := RoomLockKey(niuniu.PassWord)
//					err = gsync.GlobalTransaction(lock, f)
//					if err != nil {
//						log.Err("%s get banker failed: %v", lock, err)
//						continue
//					}
//				}
//			} else if niuniu.Status == enumniu.GameStatusSetBet && niuUser.Status == enumniu.UserStatusGetBanker {
//				f := func() error {
//					key := int32(addTime - 4)
//					err := SetBet(niuUser.UserID, key)
//					if err != nil {
//						return nil
//					}
//					m[niuniu] = []int32{enumniu.UserStatusSetBet, niuUser.UserID, key}
//					return nil
//				}
//				lock := RoomLockKey(niuniu.PassWord)
//				err = gsync.GlobalTransaction(lock, f)
//				if err != nil {
//					log.Err("%s get banker failed: %v", lock, err)
//					continue
//				}
//			} else if niuniu.Status == enumniu.GameStatusSubmitCard && niuUser.Status == enumniu.UserStatusSetBet {
//				f := func() error {
//					err := SubmitCard(niuUser.UserID)
//					if err != nil {
//						return nil
//					}
//					m[niuniu] = []int32{enumniu.UserStatusSubmitCard, niuUser.UserID}
//					return nil
//				}
//				lock := RoomLockKey(niuniu.PassWord)
//				err = gsync.GlobalTransaction(lock, f)
//				if err != nil {
//					log.Err("%s get banker failed: %v", lock, err)
//					continue
//				}
//			}
//		}
//	}
//	return m
//}

func UpdateNiuniu(niu *mdniu.Niuniu) error {

	err := cacheniu.UpdateGame(niu)
	if err != nil {
		log.Err("niuniu set session failed, %v", err)
		return err
	}
	return nil
	//f := func() error {
	//
	//}
	//
	//lock := RoomLockKey(niu.PassWord)
	//err := gsync.GlobalTransaction(lock, f)
	//if err != nil {
	//	log.Err("%s set update failed: %v", lock, err)
	//	return err
	//}
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

func GetBanker(uid int32, key int32) (*mdniu.Niuniu, error) {
	mdr, err := room.GetRoomByUserID(uid)
	if err != nil {
		return nil, err
	}
	//value, err := tools.Int2String(str)
	var value int32
	if v, ok := enumniu.BankerScoreMap[key]; !ok {
		return nil, errorsniu.ErrParam
	} else {
		value = v
	}

	//value := enumniu.BankerScoreMap[key]

	// if value < 0 || value > 4 {
	// 	return 0, errorsniu.ErrParam
	// }

	if mdr.Status > enumroom.RoomStatusStarted {
		return nil, errroom.ErrGameIsDone
	}

	if mdr.Giveup == enumroom.WaitGiveUp {
		return nil, errroom.ErrInGiveUp
	}

	niu, err := cacheniu.GetGame(mdr.RoomID)
	if niu == nil {
		return nil, errorsniu.ErrGameNoFind
	}
	if niu.Status != enumniu.GameStatusInit {
		return nil, errorsniu.ErrBankerDone
	}

	allReady, userResult := GetUserAndAllOtherStatusReady(niu, uid,
		enumniu.GetBankerStatus)

	if userResult == nil {
		return nil, errorsniu.ErrUserNotInGame
	}

	if userResult.Status > enumniu.UserStatusInit {
		return nil, errorsniu.ErrAlreadyGetBanker
	}

	userResult.Status = enumniu.UserStatusGetBanker
	userResult.Info.BankerScore = value
	if allReady {
		niu.Status = enumniu.GameStatusGetBanker
	}
	err = UpdateNiuniu(niu)
	if err != nil {
		log.Err("niuniu set session failed, %v", err)
		return nil, err
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
	return niu, nil //
}

func SetBet(uid int32, key int32) (*mdniu.Niuniu, error) {
	var value int32
	mdr, err := cacher.GetRoomUserID(uid)
	if err != nil {
		return nil, err
	}

	if key < 1 || key > 5 {
		return nil, errorsniu.ErrParam
	}
	value = key
	//if v, ok := enumniu.BetScoreMap[key]; !ok {
	//	return errorsniu.ErrParam
	//} else {
	//	value = v
	//}

	//玩家推注
	if key == enumroom.BetKeyPushOn {
		gru := &mdroom.GameUserResult{}
		for _, ru := range mdr.UserResults {
			if ru.UserID == uid {
				gru = ru
			}
		}
		//若玩家可以推注
		if gru.CanPushOn == enumroom.PushOnBet {
			value = gru.PushOnBetScore
			gru.CanPushOn = enumroom.NoPushOnBet
			gru.LastPushOnBet = mdr.RoundNow

			err = cacher.UpdateRoom(mdr)
			if err != nil {
				log.Err("niuniu update room set bet redis failed,%v | %v",
					mdr, err)
				return nil, err
			}
		} else {
			return nil, errroom.ErrPushOnBet
		}
	}

	if mdr.Status > enumroom.RoomStatusStarted {
		return nil, errroom.ErrGameIsDone
	}
	if mdr.Giveup == enumroom.WaitGiveUp {
		return nil, errroom.ErrInGiveUp
	}
	niu, err := cacheniu.GetGame(mdr.RoomID)

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
	if key == enumroom.BetKeyPushOn {
		userResult.Info.BetScore = value
	} else {
		userResult.Info.BetScore = value * niu.RoomParam.BetScore
	}

	if allReady {
		niu.Status = enumniu.GameStatusAllSetBet
	}

	err = UpdateNiuniu(niu)
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

	return niu, nil //
}

func SubmitCard(uid int32) (*mdniu.Niuniu, error) {
	mdr, err := cacher.GetRoomUserID(uid)
	if err != nil {
		return nil, err
	}
	if mdr.Status > enumroom.RoomStatusStarted {
		if mdr.Giveup == enumroom.WaitGiveUp {
			return nil, errroom.ErrInGiveUp
		}
		return nil, errroom.ErrGameIsDone
	}

	niu, err := cacheniu.GetGame(mdr.RoomID)
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
		return nil, errorsniu.ErrAlreadySubmitCard
	}

	userResult.Status = enumniu.UserStatusSubmitCard

	if allReady {
		niu.Status = enumniu.GameStatusStarted
	}

	err = UpdateNiuniu(niu)
	if err != nil {
		log.Err("niuniu set session failed, %v", err)
		return nil, err
	}
	return niu, nil //
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
				BankerScore: 1,
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
			if userResult.Info.Role != enumroom.UserRoleMaster {
				userResult.Info.BetScore = enumniu.MinSetBet * niu.RoomParam.BetScore
			}
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

func SetNextPlayerBanker(mdr *mdroom.Room, bankerLast int32) int32 {
	var bankerID int32
	for i := 0; i < len(mdr.Users); i++ {
		mdr.Users[i].Ready = enumroom.UserUnready
		if mdr.Users[i].UserID == bankerLast {
			if i == len(mdr.Users)-1 {
				mdr.Users[0].Role = enumroom.UserRoleMaster
				bankerID = mdr.Users[0].UserID
			} else {
				mdr.Users[i+1].Role = enumroom.UserRoleMaster
				bankerID = mdr.Users[i+1].UserID
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

func NiuniuRecovery(rid int32) (*mdniu.Niuniu,
	error) {
	niu, err := cacheniu.GetGame(rid)
	if err != nil {
		return nil, err
	}
	if niu == nil {
		niu, err = dbniu.GetLastNiuniuByRoomID(db.DB(), rid)
		if err != nil {
			return nil, err
		}
	}
	if niu == nil {
		return nil, errorsniu.ErrGameNotExist
	}
	return niu, nil
}

func CleanGame() error {
	var gids []int32
	rids, err := cacher.GetAllDeleteRoomKey(enumroom.NiuniuGameType)
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
			log.Debug("clean niuniu game:%d|%d|%+v\n", game.GameID, game.RoomID, game.Ids)
			gids = append(gids, game.GameID)
			err = cacheniu.DeleteGame(game)
			if err != nil {
				log.Err(" delete niuniu set session failed, %v",
					err)
				continue
			}
			err = cacher.CleanDeleteRoom(enumniu.GameID, game.RoomID)
			if err != nil {
				log.Err(" delete niuniu delete room session failed,roomid:%d,err: %v", game.RoomID,
					err)
				continue
			}
		} else {
			err = cacher.CleanDeleteRoom(enumniu.GameID, rid)
			if err != nil {
				log.Err(" delete null game niuniu delete room session failed,roomid:%d,err: %v", rid,
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

//func NiuniuRecovery(rid int32, uid int32) (int32, *mdniu.NiuniuRoomResult,
//	error) {
//	niu, err := cacheniu.GetGame(rid)
//	if err != nil {
//		return 0, nil, err
//	}
//	if niu == nil {
//		niu, err = dbniu.GetLastNiuniuByRoomID(db.DB(), rid)
//		if err != nil {
//			return 0, nil, err
//		}
//	}
//	if niu == nil {
//		return 0, nil, errorsniu.ErrGameNotExist
//	}
//	return niu.Status, niu.Result, nil
//}

func NiuniuExist(uid int32, rid int32) (*pbniu.NiuniuRecoveryReply, error) {
	out := &pbniu.NiuniuRecoveryReply{}
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
	niu, err := NiuniuRecovery(roomRecovery.Room.RoomID)
	if err != nil {
		return nil, err
	}
	out.NiuniuExist = niu.Result.ToProto()
	//out.NiuniuExist.Status = enumniu.ToGameStatusMap[niu.Status]
	for _, gur := range out.NiuniuExist.List {
		if uid != gur.UserID {
			gur.PushOnBet = 0
			if niu.Status < enumniu.GameStatusSubmitCard {
				gur.Cards = nil
			}
		}
	}
	var time int32
	switch niu.Status {
	case enumniu.GameStatusInit:
		time = enumniu.GetBankerTime
		break
	case enumniu.GameStatusGetBanker:
		time = enumniu.SetBetTime
		break
	case enumniu.GameStatusSetBet:
		time = enumniu.SetBetTime
		break
	case enumniu.GameStatusSubmitCard:
		time = enumniu.SubmitCardTime
		break
	}
	out.CountDown = &pbniu.CountDown{
		ServerTime: niu.OpDateAt.Unix(),
		Count:      time,
	}
	out.NiuniuExist.Status = enumniu.ToGameStatusMap[niu.Status]
	return out, nil
}
