package context

import (
	mdu "playcards/model/user/mod"

	"github.com/micro/go-micro/metadata"
	"golang.org/x/net/context"
)

func SetUser(ctx context.Context, u *mdu.User) context.Context {
	return context.WithValue(ctx, "user", u)
}

func GetUser(ctx context.Context) *mdu.User {
	return ctx.Value("user").(*mdu.User)
}

func ExtractMetaData(ctx context.Context) context.Context {
	m, ok := metadata.FromContext(ctx)
	if !ok {
		return ctx
	}

	for key, val := range m {
		ctx = context.WithValue(ctx, key, val)
	}
	return ctx
}

func NewContext(token string) context.Context {
	ctx := metadata.NewContext(context.Background(), map[string]string{
		"Token": token,
	})
	return ctx
}
