package handler

import (
	mdpage "playcards/model/page"
	"playcards/model/user"
	cacheuser "playcards/model/user/cache"
	mdu "playcards/model/user/mod"
	pbu "playcards/proto/user"
	"playcards/utils/auth"
	"playcards/utils/log"
	utilpb "playcards/utils/proto"

	"golang.org/x/net/context"
)

type UserSrv struct {
}

func (us *UserSrv) Register(ctx context.Context, req *pbu.User,
	rsp *pbu.RegisterReply) error {

	req.Rights = auth.RightsPlayer

	uid, err := user.Register(mdu.UserFromProto(req))
	if err != nil {
		return err
	}

	rsp.UserID = uid
	log.Debug("register user: %d", uid)

	return nil
}

func (us *UserSrv) AddUser(ctx context.Context, req *pbu.User,
	rsp *pbu.RegisterReply) error {

	uid, err := user.Register(mdu.UserFromProto(req))
	if err != nil {
		return err
	}

	rsp.UserID = uid
	log.Debug("Add user: %d", uid)

	return nil
}

func (us *UserSrv) Login(ctx context.Context, req *pbu.User,
	rsp *pbu.LoginReply) error {

	u, err := user.Login(mdu.UserFromProto(req))
	if err != nil {
		return err
	}

	token, err := cacheuser.SetUser(u)
	if err != nil {
		log.Err("user login set session failed, %v", err)
		return err
	}
	rsp.Token = token
	log.Debug("login: %v", u)

	return nil
}

func (us *UserSrv) UserInfo(ctx context.Context, req *pbu.UserInfoReq,
	rsp *pbu.User) error {

	u, err := auth.GetUser(ctx)
	if err != nil {
		return err
	}

	*rsp = *u.ToProto()

	return nil
}

func (u *UserSrv) PageUserList(ctx context.Context,
	req *pbu.PageUserListRequest, rsp *pbu.PageUserListReply) error {

	page := mdpage.PageOptionFromProto(req.Page)
	l, rows, err := user.PageUserList(page, mdu.UserFromPageRequestProto(req))
	if err != nil {
		return err
	}

	err = utilpb.ProtoSlice(l, &rsp.List)
	if err != nil {
		return err
	}

	rsp.Count = rows
	return nil
}

func (us *UserSrv) UpdateUser(ctx context.Context, req *pbu.User,
	rsp *pbu.User) error {

	_, err := user.UpdateUser(mdu.UserFromProto(req))
	if err != nil {
		return err
	}
	return nil
}
