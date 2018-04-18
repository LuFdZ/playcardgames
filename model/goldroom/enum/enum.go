package enum

import (
	mdgr "playcards/model/goldroom/mod"
)

const (
	GoldRoomLevelPrimary = 1
	GoldRoomLevelMiddle  = 2
)

//globle config
const (
	LoopTime               = 500
	MaxRecordCount         = 10
	RoomDelayMinutes       = 5.0
	RoomCodeMax            = 999999
	RoomCodeMin            = 100000
	RoomGiveupCleanMinutes = 5.0
	ShuffleDelaySeconds    = 3
)

const (
	NoFull = 1
	Full   = 2
)

const (
	Player = 1
	Robot  = 2
)

const (
	Sit   = 1
	UnSit = 2
)

const (
	JoinGame   = 1
	UnJoinGame = 2
)

const (
	UserOpJoin   = 110 //机器人操作值 加入游戏
	UserOpReady  = 120 //机器人操作值 游戏准备
	UserOpRemove = 130 //清除不活动玩家
)

//make(map[int32]map[int32][]int32)
//金币场配置表
//map[游戏类型ID]{map[游戏场等级ID]{底分，入场条件，收费}}
var GoldRoomCostMap = map[int32]map[int32][]int64{
	1001: {1: []int64{100, 20000, 100}, 2: []int64{500, 100000, 100}},
	1002: {1: []int64{100, 20000, 100}, 2: []int64{500, 100000, 100}},
	1003: {1: []int64{100, 20000, 100}, 2: []int64{500, 100000, 100}},
	1004: {1: []int64{100, 20000, 100}, 2: []int64{500, 100000, 100}},
}

var GoldRoomInfoMap = map[int32]mdgr.GoldRoomInfo{
	1001: mdgr.GoldRoomInfo{GameType: 1001, MaxNumber: 4, MinNumber: 2, RoomParam: "{\"BankerAddScore\":0,\"Time\":30,\"Joke\":0,\"Times\":2}",},
	1002: mdgr.GoldRoomInfo{GameType: 1002, MaxNumber: 6, MinNumber: 2, RoomParam: "{\"Times\":2,\"BankerType\":1,\"BetScore\":1,\"AdvanceOptions\":[\"0\",\"0\"],\"SpecialCards\":[\"1\",\"0\",\"0\",\"0\",\"1\",\"0\",\"1\"]}",},
	1004: mdgr.GoldRoomInfo{GameType: 1003, MaxNumber: 4, MinNumber: 2, RoomParam: "{\"ScoreType\":2,\"BetType\":2}",},
}
