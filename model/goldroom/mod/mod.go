package mod

//type GoldRoom struct {
//	RoomID      int32
//	GameType    int32
//	Status      int32
//	MaxNumber   int32
//	//NumberNow   int32
//	Level       int32
//	UserList    []*GoldRoomPlayer
//	OpTime      time.Time
//}
//
//type GoldRoomPlayer struct {
//	UserID int32
//	Type   int32     // 1普通玩家 2机器人
//	Status int32     // 1坐下 2没坐下
//	OpTime time.Time //等待时间
//}

type GoldRoomInfo struct {
	GameType  int32  //游戏类型
	MaxNumber int32  //游戏最大参与人数
	MinNumber int32  //游戏最小开始人数
	RoomParam string //游戏参数
}
