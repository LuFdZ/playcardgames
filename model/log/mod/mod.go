package mod

import (
	pblog "playcards/proto/log"
	"time"
)

type ClientErrorLog struct {
	ID            int32 `gorm:"primary_key"`
	UserID        int32
	ClientAddress string
	Message       string
	Condition     string
	StackTrace    string
	SystemInfo    string
	CreatedAt     *time.Time
}

func ClientErrorLogFromProto(l *pblog.ClientErrorLog) *ClientErrorLog {
	return &ClientErrorLog{
		Message:    l.Message,
		Condition:  l.Condition,
		StackTrace: l.StackTrace,
		SystemInfo: l.SystemInfo,
	}
}

type ErrLog struct {
	ServerCode int32
	Date       *time.Time
	Error      string
	Times      int32
}
