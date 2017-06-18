package handler

import (
	"playcards/model/room"
	pbr "playcards/proto/room"
	"playcards/utils/auth"

	"golang.org/x/net/context"
)

type RoomSrv struct {
}

func (rs *RoomSrv) CreateRoom(ctx context.Context, req *pbr.CreateRoomRequest,
	rsp *pbr.Room) error {
	u, err := auth.GetUser(ctx)
	if err != nil {
		return err
	}
	r, err := room.CreateRoom(req.Password, req.GameType, req.PlayerMaxNum,
		u.UserID)
	if err != nil {
		return err
	}
	*rsp = *r.ToProto()
	return nil
}

func (rs *RoomSrv) JoinRoom(ctx context.Context, req *pbr.JoinRoomRequest,
	rsp *pbr.Room) error {
	u, err := auth.GetUser(ctx)
	if err != nil {
		return err
	}
	r, err := room.JoinRoom(req.Password, u.UserID)
	if err != nil {
		return err
	}
	*rsp = *r.ToProto()
	return nil
}
