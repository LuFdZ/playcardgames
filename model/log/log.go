package log

import (
	dblog "playcards/model/log/db"
	mdlog "playcards/model/log/mod"
	cachelog "playcards/model/log/cache"
	enumlog "playcards/model/log/enum"
	"playcards/utils/log"
	"net/http"
	"playcards/utils/db"
	"strings"
)

func AddClientErrorLog(l *mdlog.ClientErrorLog) error {
	return dblog.AddClientErrorLog(db.DB(), l)
}

func GetAllErrLog() {
	keys, values := cachelog.GetAllErrLog()
	//fmt.Printf("AAAAAGetAllErrLog:%+v\n", keys)
	if len(keys) == 0 {
		return
	}
	for i := 0; i < len(keys); i++ {
		SendErrLog(keys[i], values[i])
	}
}

func SendErrLog(key string, value string) {
	requestLine := strings.Join([]string{enumlog.Url, "?type=",
		enumlog.Server, "&ver=", enumlog.Version, "&cnt=", value}, "")
	//fmt.Printf("SendErrLog:%s\n", requestLine)
	rsp, err := http.Get(requestLine)
	log.Info("SendErrLog rsp:%s,url:%s",rsp,requestLine)
	if err != nil {
		log.Err("send err log fail,value:%s,rsp:%s,err:%s", value,rsp, err)
	}
	cachelog.DeleteGame(key)

}
