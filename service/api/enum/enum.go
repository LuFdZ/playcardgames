package enum

import "playcards/utils/errors"

const (
	VERSION = "bcr.1.0.0"
)

var (
	APIServiceName      = "bcr.api"
	Namespace           = "bcr.srv"
	UserServiceName     = "bcr.srv.user"
	BillServiceName     = "bcr.srv.bill"
	FriendServiceName   = "bcr.srv.friend"
	RegionServiceName   = "bcr.srv.region"
	ChatServiceName     = "bcr.srv.chat"
	StoreServiceName    = "bcr.srv.store"
	WebServiceName      = "bcr.srv.web"
	ConfigServiceName   = "bcr.srv.config"
	ZodiacServiceName   = "bcr.srv.zodiac"
	ActivityServiceName = "bcr.srv.activity"
	LogServiceName      = "bcr.srv.log"
)

var (
	ErrInvalidPing      = errors.BadRequest(30001, "bad ping")
	ErrDecryptID        = errors.BadRequest(30002, "invlid id")
	ErrReadBodyFailed   = errors.BadRequest(30003, "read body failed")
	ErrDecodeArgsFailed = errors.BadRequest(30004, "decode args failed")
)
