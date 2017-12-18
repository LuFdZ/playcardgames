package main

import (
	"fmt"
	"time"
	//"encoding/json"
	//"strings"
	//"math"
	"github.com/yuin/gopher-lua"
)

func main() {

	l := lua.NewState()
	defer l.Close()
	var err error
	if err = l.DoFile("/Users/lufdz/GoPro/src/playcards/lua/doudizhulua/Logic.lua"); err != nil {
		fmt.Printf("doudizhu logic do file %+v", err)
	}

	ostime := time.Now().UnixNano()
	if err = l.DoString(fmt.Sprintf("return G_Init(%d)", ostime)); err != nil {
		fmt.Printf("doudizhu G_Init error %+v", err)
		return
	}

	if err := l.DoString("return G_Reset()"); err != nil {
		fmt.Printf("doudizhu G_Reset %+v", err)
		return
	}

	if err := l.DoString(fmt.Sprintf("return G_GetCards()")); err != nil {
		fmt.Printf("doudizhu return logic get cards %v", err)
		return
	}
	getCards := l.Get(-1)
	l.Pop(1)
	TotalMap := getCards.(*lua.LTable)
	TotalMap.ForEach(func(key lua.LValue, value lua.LValue) {
		CardsList := value.(*lua.LTable)
		CardsList.ForEach(func(key lua.LValue, value lua.LValue){

		})
	})

}
