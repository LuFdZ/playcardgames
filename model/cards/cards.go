package cards

import (
	"fmt"
	"math/rand"
	"time"
)

type CardCenter struct {
	CardsID      int32
	CardsNum     int32
	DeckNum      int32
	HasKings     int32
	ShuffleAways bool
	CardsList    []Card
}

type Card struct {
	CardID int32
	Name   string
	Value  int32
	Color  int32
}

func (cc *CardCenter) Init() {
	for i := 0; i < (int)(cc.DeckNum); i++ {

	}
}

func (cc *CardCenter) Shuffle() {
	vals := cc.CardsList
	r := rand.New(rand.NewSource(time.Now().Unix()))
	for _, i := range r.Perm(len(vals)) {
		val := vals[i]
		fmt.Println(val)
	}
	for index := 0; index < len(cc.CardsList); index++ {
		fmt.Printf("arr[%d]=%d \n", index, cc.CardsList[index].CardID)
	}
}
