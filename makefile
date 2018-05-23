GO = go
GOGET = ${GO} get
GOINSTALL = ${GO} install
BUILD = ${GO} build ${GOFLAGS}
MIGRATE = ${GOPATH}/bin/migrate
MIGRATE_URL = -url "mysql://root:@tcp(127.0.0.1:3306)/playcards?parseTime=true" -path ./sql
MIGRATE_CMD = ${MIGRATE} ${MIGRATE_URL}
PROTOC = protoc
PROTO_BUILD = ${PROTOC} -I.. -I. --go_out=plugins=micro:.

all:  api-srv           \
      user-srv          \
      bill-srv          \
      web-srv           \
      config-srv        \
      log-srv           \
      activity-srv      \
      room-srv          \
      thirteen-srv      \
      notice-srv        \
      niuniu-srv        \
      club-srv          \
      common-srv        \
      doudizhu-srv      \
      fourcard-srv      \
      mail-srv          \
      twocard-srv       \
      runcard-srv       \

dev:
	-${GOGET} github.com/golang/lint
	-${GOINSTALL} github.com/golang/lint
	-${GOGET} code.google.com/p/go.tools/cmd/oracle

	-${GOGET} golang.org/x/tools/cmd/goimports

	-${GOGET} github.com/nsf/gocode

	-${GOGET} github.com/rogpeppe/godef
	-${GOGET} github.com/mattes/migrate
	-${GOGET} golang.org/x/tools/cmd/godoc
	-${GOGET} golang.org/x/tools/cmd/goimports
	-${GOGET} golang.org/x/tools/cmd/gotype
	-${GOGET} golang.org/x/tools/cmd/cover
	-${GOGET} golang.org/x/tools/cmd/gorename
	-${GOGET} golang.org/x/tools/cmd/oracle
	-${GOGET} golang.org/x/tools/cmd/vet
dep:
	-${GOGET} github.com/micro/protobuf/proto
	-${GOGET} github.com/micro/protobuf/protoc-gen-go
	-${GOGET} github.com/micro/micro
	-${GOGET} github.com/micro/go-micro
	-${GOGET} github.com/micro/go-os
	-${GOGET} github.com/micro/config-srv/proto/config
	-${GOGET} github.com/micro/go-plugins
	-${GOGET} github.com/micro/go-web
	-${GOGET} github.com/nats-io/nats
	-${GOGET} github.com/jinzhu/gorm
	-${GOGET} github.com/jinzhu/now
	-${GOGET} gopkg.in/fsnotify.v1
	-${GOGET} gopkg.in/redis.v5
	-${GOGET} github.com/bitly/go-simplejson
	-${GOGET} github.com/twinj/uuid
	-${GOGET} github.com/imdario/mergo
	-${GOGET} github.com/go-sql-driver/mysql
	-${GOGET} github.com/Masterminds/squirrel
	-${GOGET} github.com/fatih/structs
	-${GOGET} github.com/nats-io/nats
	-${GOGET} github.com/asaskevich/govalidator
	-${GOGET} gopkg.in/go-playground/validator.v8
	-${GOGET} github.com/rs/cors
	-${GOGET} github.com/shopspring/decimal
	-${GOGET} github.com/google/gops
	-${GOGET} github.com/yuin/gopher-lua

gen: dep
	${PROTO_BUILD} ./proto/page/page.proto
	${PROTO_BUILD} ./proto/user/user.proto
	${PROTO_BUILD} ./proto/bill/bill.proto
	${PROTO_BUILD} ./proto/time/time.proto
	${PROTO_BUILD} ./proto/web/web.proto
	${PROTO_BUILD} ./proto/config/config.proto
	${PROTO_BUILD} ./proto/activity/activity.proto
	${PROTO_BUILD} ./proto/log/log.proto
	${PROTO_BUILD} ./proto/room/room.proto
	${PROTO_BUILD} ./proto/thirteen/thirteen.proto
	${PROTO_BUILD} ./proto/notice/notice.proto
	${PROTO_BUILD} ./proto/niuniu/niuniu.proto
	${PROTO_BUILD} ./proto/club/club.proto
	${PROTO_BUILD} ./proto/common/common.proto
	${PROTO_BUILD} ./proto/doudizhu/doudizhu.proto
	${PROTO_BUILD} ./proto/fourcard/fourcard.proto
	${PROTO_BUILD} ./proto/mail/mail.proto
	${PROTO_BUILD} ./proto/goldroom/goldroom.proto
	${PROTO_BUILD} ./proto/twocard/twocard.proto
	${PROTO_BUILD} ./proto/runcard/runcard.proto


api-srv: gen
	${BUILD} -o ./bin/api-srv service/api/main.go

web-srv: gen
	${BUILD} -o ./bin/web-srv service/web/main.go

user-srv: gen
	${BUILD} -o ./bin/user-srv service/user/main.go

bill-srv: gen
	${BUILD} -o ./bin/bill-srv service/bill/main.go

config-srv: gen
	${BUILD} -o ./bin/config-srv service/config/main.go

activity-srv: gen
	${BUILD} -o ./bin/activity-srv service/activity/main.go

log-srv: gen
	${BUILD} -o ./bin/log-srv service/log/main.go

room-srv: gen
	${BUILD} -o ./bin/room-srv service/room/main.go

thirteen-srv: gen
	${BUILD} -o ./bin/thirteen-srv service/thirteen/main.go

notice-srv: gen
	${BUILD} -o ./bin/notice-srv service/notice/main.go

niuniu-srv: gen
	${BUILD} -o ./bin/niuniu-srv service/niuniu/main.go

club-srv: gen
	${BUILD} -o ./bin/club-srv service/club/main.go

common-srv: gen
	${BUILD} -o ./bin/common-srv service/common/main.go

doudizhu-srv: gen
	${BUILD} -o ./bin/doudizhu-srv service/doudizhu/main.go

fourcard-srv: gen
	${BUILD} -o ./bin/fourcard-srv service/fourcard/main.go
mail-srv: gen
	${BUILD} -o ./bin/mail-srv service/mail/main.go
goldroom-srv: gen
	${BUILD} -o ./bin/goldroom-srv service/goldroom/main.go
twocard-srv: gen
	${BUILD} -o ./bin/twocard-srv service/twocard/main.go
runcard-srv: gen
	${BUILD} -o ./bin/runcard-srv service/runcard/main.go

db-reset-init:
	${MIGRATE_CMD} goto 1
	${MIGRATE_CMD} reset
	${MIGRATE_CMD} version

.PHONY: all gen dep



