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
	utilproto "playcards/utils/proto"
	"strings"

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
	address, _ := ctx.Value("X-Real-Ip").(string)
	u, diamond, err := user.Login(mdu.UserFromProto(req), address)
	if err != nil {
		return err
	}

	token, err := cacheuser.SetUser(u)
	if err != nil {
		log.Err("user login set session failed, %v", err)
		return err
	}
	rsp.Token = token
	reply := u.ToProto()
	reply.Diamond = diamond
	rsp.User = reply
	log.Debug("login: %v", u)
	//room.AutoSubscribe(u.UserID)
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

func (us *UserSrv) GetUser(ctx context.Context, req *pbu.User,
	rsp *pbu.User) error {

	_, err := auth.GetUser(ctx)
	if err != nil {
		return err
	}

	um, err := user.GetUser(mdu.UserFromProto(req))
	*rsp = *um.ToProto()
	return nil
}

func (us *UserSrv) PageUserList(ctx context.Context,
	req *pbu.PageUserListRequest, rsp *pbu.PageUserListReply) error {
	_, err := auth.GetUser(ctx)
	if err != nil {
		return err
	}
	rsp.Result = 2
	rsp.Code = 101
	page := mdpage.PageOptionFromProto(req.Page)
	l, rows, err := user.PageUserList(page, mdu.UserFromPageRequestProto(req))
	if err != nil {
		rsp.Code = 102
		return err
	}

	err = utilpb.ProtoSlice(l, &rsp.List)
	if err != nil {
		rsp.Code = 103
		return err
	}
	if len(rsp.List) > 0 {
		rsp.Code = 0
		rsp.Result = 1
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

func (us *UserSrv) GetToken(ctx context.Context, req *pbu.GetTokenRequest,
	rsp *pbu.GetTokenReply) error {
	rsp.Result = 2
	rsp.Code = 101
	token, _ := cacheuser.GetUserByID(req.UserID)
	if len(token) > 0 {
		rsp.Result = 1
		rsp.Code = 0
		rsp.Token = token
	}
	return nil
}

func (us *UserSrv) CheckUser(ctx context.Context, req *pbu.CheckUserRequest,
	rsp *pbu.CheckUserReply) error {
	rsp.Result = 2
	rsp.Code = 101
	token, user := cacheuser.GetUserByID(req.UserID)
	if user != nil || len(token) > 0 {
		if token == req.Token {
			rsp.Code = 0
			rsp.Result = 1
		}
	}
	return nil
}

func (us *UserSrv) WXLogin(ctx context.Context, req *pbu.WXLoginRequest,
	rsp *pbu.WXLoginRsply) error {
	forwarded := ctx.Value("X-Forwarded-For")
	address := ""
	if forwarded != nil {
		addresslist, _ := forwarded.(string) //ctx.Value("X-Real-Ip").(string)
		list := strings.Split(addresslist, ",")
		address = list[0]
	}
	diamond, u, err := user.WXLogin(mdu.UserFromWXLoginRequestProto(req), req.Code, address)
	if err != nil {
		return err
	}
	u.MobileOs = req.MobileOs
	u.Version = req.Version
	u.Channel = req.Channel
	token, err := cacheuser.SetUser(u)
	if err != nil {
		log.Err("user login set session failed, %v", err)
		return err
	}
	rsp.Token = token
	log.Debug("login: %v|%s", u, address)
	reply := u.ToProto()
	reply.Diamond = diamond
	rsp.User = u.ToProto()
	//fmt.Printf("WXLogin:%v\n",rsp.User)
	return nil
}

func (us *UserSrv) DayActiveUserList(ctx context.Context, req *pbu.GetUserDayActiveRequest,
	rsp *pbu.GetUserDayActiveRsply) error {
	_, err := auth.GetUser(ctx)
	if err != nil {
		return err
	}
	ul, pr := user.DayActiveUserList(req.Page)
	resply := &pbu.GetUserDayActiveRsply{
		PageReply: pr.ToProto(),
	}
	utilproto.ProtoSlice(ul, &resply.List)
	*rsp = *resply
	return nil
}

func (us *UserSrv) GetUserCount(ctx context.Context, req *pbu.User,
	rsp *pbu.ConutRsply) error {
	_, err := auth.GetUser(ctx)
	if err != nil {
		return err
	}
	resply := &pbu.ConutRsply{}
	count, err := user.GetUserOnlineCount()
	if err != nil {
		return err
	}
	resply.OnlineCount = count
	count = user.DayActiveUserCount()
	resply.ActiveCount = count
	count = user.DayNewUserCount()
	resply.NewCount = count
	*rsp = *resply
	return nil
}

func (us *UserSrv) SetLocation(ctx context.Context, req *pbu.SetLocationRequest,
	rsp *pbu.UserRsply) error {
	u, err := auth.GetUser(ctx)
	if err != nil {
		return err
	}
	resply := &pbu.UserRsply{
		Result: 1,
	}
	err = user.SetLocation(u, req.Json)
	if err != nil {
		return err
	}

	*rsp = *resply
	return nil
}

func (us *UserSrv) RefreshUserCount(ctx context.Context, req *pbu.SetLocationRequest,
	rsp *pbu.UserRsply) error {
	resply := &pbu.UserRsply{
		Result: 1,
	}
	err := user.RefreshUserCount()
	if err != nil {
		return err
	}
	*rsp = *resply
	return nil
}