package log

import (
	dblog "playcards/model/log/db"
	mdlog "playcards/model/log/mod"
	"playcards/utils/db"
)

func AddClientErrorLog(l *mdlog.ClientErrorLog) error {
	return dblog.AddClientErrorLog(db.DB(), l)
}
