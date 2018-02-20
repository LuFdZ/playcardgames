package enum

import "playcards/utils/errors"

const (
	VERSION = "playcards.1.0.0"
)

var (
	APIServiceName      = "playcards.api"
	Namespace           = "playcards.srv"
	UserServiceName     = "playcards.srv.user"
	BillServiceName     = "playcards.srv.bill"
	WebServiceName      = "playcards.srv.web"
	ConfigServiceName   = "playcards.srv.config"
	ActivityServiceName = "playcards.srv.activity"
	LogServiceName      = "playcards.srv.log"
	RoomServiceName     = "playcards.srv.room"
	ThirteenServiceName = "playcards.srv.thirteen"
	NoticeServiceName   = "playcards.srv.notice"
	NiuniuServiceName   = "playcards.srv.niuniu"
	ClubServiceName     = "playcards.srv.club"
	CommonServiceName   = "playcards.srv.common"
	DoudizhuServiceName = "playcards.srv.doudizhu"
	FourCardServiceName = "playcards.srv.fourcard"
	MailServiceName     = "playcards.srv.mail"
	GoldRoomServiceName = "playcards.srv.goldroom"
)

var (
	ErrInvalidPing      = errors.BadRequest(30001, "bad ping")
	ErrDecryptID        = errors.BadRequest(30002, "invlid id")
	ErrReadBodyFailed   = errors.BadRequest(30003, "read body failed")
	ErrDecodeArgsFailed = errors.BadRequest(30004, "decode args failed")
)
