package enum

const (
	ErrUID             = -1
	Salt               = "mChkzskxqc7biVQtLvPjyQkha4c"
	RegisterBalance    = 0
	NewUserDiamond     = 100
	NewUserGold        = 0
	MaxUserRecordCount = 20
)

const (
	UserTableName     = "users"
	PropertyTableName = "properties"
)

const (
	AppId           = "wx6bcafc0ab9c4478f"
	Secret          = "9008b23c88159a18a1b252c1ad7c73bc"
	GetTokenUrl     = "https://api.weixin.qq.com/sns/oauth2/access_token"
	CheckTokenUrl   = "https://api.weixin.qq.com/sns/auth"
	RefreshTokenUrl = "https://api.weixin.qq.com/sns/oauth2/refresh_token"
	GetUserUrl      = "https://api.weixin.qq.com/sns/userinfo"
)

const (
	ResultStatusSuccess = 1
	ResultStatusFail    = 2
)

const (
	UserOnline = 1
	UserUnline = 2
)
