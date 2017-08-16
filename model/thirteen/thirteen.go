package thirteen

import (
	"fmt"
	dbbill "playcards/model/bill/db"
	mdbill "playcards/model/bill/mod"
	cacheroom "playcards/model/room/cache"
	dbr "playcards/model/room/db"
	enumr "playcards/model/room/enum"
	mdr "playcards/model/room/mod"
	dbt "playcards/model/thirteen/db"
	enumt "playcards/model/thirteen/enum"
	mdt "playcards/model/thirteen/mod"
	"playcards/utils/db"
	"playcards/utils/log"
	"strconv"

	"github.com/jinzhu/gorm"
	"github.com/yuin/gopher-lua"
)

func CreateThirteen() {
	rooms, err := GetRoomsByStatusAndGameType()
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
		if err := l.DoString("return Logic:GetCards()"); err != nil {
			log.Err("thirteen logic do string %v", err)
		}
		getCards := l.Get(1)
		l.Pop(1)
		var groupCards []*mdt.GroupCard
		//for _, user := range room.Users {
		for i := 0; i < 4; i++ {
			var userID int32
			var role int32
			if i+1 > len(room.Users) {
				userID = -1
				role = -1
			} else {
				userID = room.Users[i].UserID
				userID = room.Users[i].Role
			}

			var cardList []*mdt.Card
			if cardsMap, ok := getCards.(*lua.LTable); ok {
				cardsMap.ForEach(func(key lua.LValue, value lua.LValue) {
					if cards, ok := value.(*lua.LTable); ok {
						var cardType int32
						var cardValue int32
						cards.ForEach(func(k lua.LValue, v lua.LValue) {
							value, _ := strconv.ParseInt(v.String(), 10, 32)
							if k.String() == "_type" {
								cardType = int32(value)
							} else {
								cardValue = int32(value)
							}
						})
						card := mdt.Card{
							Type:  int32(cardType),
							Value: int32(cardValue),
						}
						cardList = append(cardList, &card)
					} else {
						log.Err("thirteen cardsMap value err %v", value)
					}
				})
				groupCard := mdt.GroupCard{
					UserID:   userID,
					Type:     role,
					Weight:   0,
					CardList: cardList,
				}
				groupCards = append(groupCards, &groupCard)
			} else {
				log.Err("thirteen cardsMap err %v", cardsMap)
			}
		}

		groupCardList := mdt.GroupCardList{
			List: groupCards,
		}
		fmt.Printf("AAAAAAA %v", groupCardList)
		thirteen := &mdt.Thirteen{
			RoomID:  room.RoomID,
			Status:  enumt.GameStatusInit,
			Index:   1,
			GameLua: l,
			Cards:   &groupCardList,
		}
		f := func(tx *gorm.DB) error {
			err := dbbill.GainBalance(tx, room.Users[0].UserID,
				&mdbill.Balance{0, 0, enumt.GameCost},
				enumbill.JournalTypeZodiac, int64(zb.ID))
			if err != nil {
				return err
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
	}

}

func SubmitCardList(uid int32, head mdt.CardList, middle mdt.CardList,
	tail mdt.CardList) {

}

func Surrender() {}

func GetGameResult() {}

func GetRoomsByStatusAndGameType() ([]*mdr.Room, error) {
	var (
		rooms []*mdr.Room
	)
	f := func(tx *gorm.DB) error {
		list, err := dbr.GetRoomsByStatusAndGameType(db.DB(),
			enumr.RoomStatusAllReady, enumt.GameID)
		if err != nil {
			return err
		}
		rooms = list
		return nil
	}
	err := db.Transaction(f)
	if err != nil {
		return nil, err
	}
	return rooms, nil
}
