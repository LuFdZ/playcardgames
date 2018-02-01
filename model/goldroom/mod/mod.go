package mod

type GoldRoom struct {
	RoomID      int32
	GameType    int32
	Status      int32
	MaxNumber   int32
	RoundNumber int32
	OpTime      int64
	UserList    []*GoldRoomPlayer
}

type GoldRoomPlayer struct {
	UserID      int32
	Type        int32
	Status      int32
	WaitingTime int64
}

