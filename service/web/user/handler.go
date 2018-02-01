package user

import (
	pbuser "playcards/proto/user"
	srvuser "playcards/service/user/handler"
	"playcards/service/web/enum"
	"playcards/service/web/clients"
	"playcards/service/web/request"
	"playcards/utils/auth"
	"encoding/json"
	gctx "playcards/utils/context"
	"time"
)

var UserEvent = []string{
	srvuser.TopicHeartbeatTimeout,
}

func ClinetHearbeatMessage(c *clients.Client, req *request.Request) error {
	var ctime int64 = 0

	err := json.Unmarshal(req.Args, &ctime)
	if err != nil {
		ctime = 0
	}
	msg := &pbuser.HeartBeat{}
	if ctime != 0 {
		msg.Stime = time.Now().Unix()
		msg.Ctime = ctime
	}
	c.SendMessage("", enum.MsgHeartbeat, msg, enum.MsgSubscribeCode)
	if c.Heartbeat(){
		ctx := gctx.NewContext(c.Token())
		rpc.Heartbeat(ctx, &pbuser.HeartbeatRequest{})
	}

	return nil
}

//func HeartbeatCallback(c *clients.Client) {
//	ctx := gctx.NewContext(c.Token())
//	rpc.Heartbeat(ctx, &pbuser.HeartbeatRequest{})
//}

func init() {
	request.RegisterHandler("ClientHeartbeat", auth.RightsPlayer,
		ClinetHearbeatMessage)
	//request.RegisterHeartbeatHandler(HeartbeatCallback)
}
