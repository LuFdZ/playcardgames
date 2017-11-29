package config

import (
	"time"

	"github.com/micro/go-os/config"
)

var ConfigDefault config.Config

func Init(c config.Config) {
	ConfigDefault = c
}

func Debug() bool {
	// TODO: default is false
	return ConfigDefault.Get("debug").Bool(true)
}

func LogLevel() string {
	return ConfigDefault.Get("log_level").String("debug")
}

func DBURL() string {
	return ConfigDefault.Get("db_url").
		String("root:@tcp(127.0.0.1:3306)/playcards?parseTime=true&loc=Asia%2FChongqing")
		//String("root:Playcards#18@tcp(127.0.0.1:3306)/playcards?parseTime=true&loc=Asia%2FChongqing")
		//String("root:Playcards#18@tcp(172.19.90.246:3306)/playcards?parseTime=true&loc=Asia%2FChongqing")
}

func RedisHost() []string {
	return ConfigDefault.Get("redis_host").
		StringSlice([]string{"127.0.0.1:6379",""})
		//StringSlice([]string{"172.19.90.246:6379"})
}

func APIHost() string {
	return ConfigDefault.Get("api_host").String("0.0.0.0:8080")
}

func WebHost() string {
	return ConfigDefault.Get("web_host").String("0.0.0.0:8999")
}

func ApiAllowedOrigins() []string {
	return ConfigDefault.Get("api_allowed_origins").StringSlice([]string{"*"})
}

func RegisterTTL() (time.Duration, time.Duration) {
	ttl := ConfigDefault.Get("register_ttl").Duration(30 * time.Second)
	interval := ttl / 2
	return ttl, interval
}
