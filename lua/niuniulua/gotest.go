 package main
import (
	"fmt"

	lua "github.com/yuin/gopher-lua"
)

func main() {
	//fmt.Printf("hello Golang\n");

	l := lua.NewState()
	defer l.Close()
	if err := l.DoFile("Logic.lua"); err != nil {
		fmt.Printf("DoFile:%v", err)
	}
	if err := l.DoString("return Logic:new()"); err != nil {
		fmt.Printf("DoString new:%v", err)
	}

	if err := l.DoString("return Logic:GetGroupTypeTest()"); err != nil {
		fmt.Printf("DoString GetGroupTypeTest:%v", err)
	}
	logic := l.Get(-1)
	fmt.Printf("getcard: %v", logic)

	// if err := l.DoString("return Logic:GetCards()"); err != nil {
	// 	fmt.Printf("DoString GetCards:%v", err)
	// }
	// logic := l.Get(-1)
	// fmt.Printf("getcard: %v", logic)

	// test := "{'RoomID':152,'List':[{'UserID':100002,'Status':4,'Info':{'BankerScore':2,'BetScore':15,'Role':2},'Cards':{'CardType':'0','CardList':['3_13','2_8','3_11','3_12','2_3']},'Score':0},{'UserID':100003,'Status':4,'Info':{'BankerScore':2,'BetScore':0,'Role':1},'Cards':{'CardType':'0','CardList':['4_6','1_6','3_9','2_5','2_10']},'Score':0}]}"
	
	// if err := l.DoString(fmt.Sprintf("return Logic:CalculateRes(\"%s\", '%s')", test, "{\"Times\":3,\"BankerType\":1}")); err != nil {
	// 	fmt.Printf("DoString CalculateRes:%v", err)
	// }
	// logic := l.Get(-1)

	// fmt.Printf("-----------------------------------------\n%v", logic)
}
