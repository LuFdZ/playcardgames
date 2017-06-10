package auth

import (
	cacheuser "playcards/model/user/cache"
	mdu "playcards/model/user/mod"
	"playcards/utils/auth/errors"
	gctx "playcards/utils/context"
	"playcards/utils/log"
	"time"

	"github.com/micro/go-micro/metadata"
	"github.com/micro/go-micro/server"
	"golang.org/x/net/context"
)

func GetToken(ctx context.Context) string {
	md, ok := metadata.FromContext(ctx)
	if !ok {
		return ""
	}
	return md["Token"]
}

func GetUser(ctx context.Context) (*mdu.User, error) {
	token := GetToken(ctx)
	if token == "" {
		return nil, errors.ErrNoToken
	}
	return GetUserByToken(token)
}

func GetUserIfExist(ctx context.Context) (*mdu.User, error) {
	token := GetToken(ctx)
	if token == "" {
		return nil, nil
	}
	return GetUserByToken(token)
}

func GetUserByToken(token string) (*mdu.User, error) {
	return cacheuser.GetUser(token)
}

func Check(urights, srights int32) error {
	if urights&srights == RightsNone {
		return errors.ErrNoPermission
	}
	return nil
}

func ServerAuth(ctx context.Context, rights int32) (*mdu.User, error) {
	u, err := GetUserIfExist(ctx)
	if err != nil {
		return nil, err
	}

	if rights == RightsNone {
		return u, nil
	}

	if u == nil {
		return nil, errors.ErrInvalidToken
	}

	if err := Check(u.Rights, rights); err != nil {
		return u, err
	}

	return u, err
}

func ServerAuthWrapper(rights map[string]int32) server.HandlerWrapper {
	return func(fn server.HandlerFunc) server.HandlerFunc {
		return func(ctx context.Context, req server.Request,
			rsp interface{}) error {

			method := req.Method()
			r, ok := rights[method]
			if !ok {
				log.Warn("no %v method found", method)
				return errors.ErrNoMethod
			}

			u, err := ServerAuth(ctx, r)
			if err != nil {
				log.Warn("%v try to call %v auth failed %v",
					u, method, err)
				return err
			}

			ctx = gctx.ExtractMetaData(ctx)
			ctx = gctx.SetUser(ctx, u)
			start := time.Now()
			err = fn(ctx, req, rsp)
			end := time.Now()

			log.Info("%v call %v, req:%v, rsp:%v, err: %v, [%v]",
				u, method, req.Request(), rsp, err, end.Sub(start))

			return err
		}
	}
}
