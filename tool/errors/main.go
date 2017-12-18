package main

import (
	"fmt"
	_ "playcards/model/activity/errors"
	_ "playcards/model/bill/errors"
	_ "playcards/model/time/errors"
	_ "playcards/model/user/errors"
	_ "playcards/service/api/enum"
	_ "playcards/utils/auth/errors"
	"playcards/utils/errors"
)

func main() {
	for _, e := range errors.Errors {
		fmt.Println(e)
	}
}
