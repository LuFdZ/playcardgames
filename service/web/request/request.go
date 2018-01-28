package request

import (
	"encoding/json"
	"fmt"
	"playcards/service/web/clients"
	"playcards/utils/log"
)

type Request struct {
	Method string
	Args   json.RawMessage
}

type RequestHandler func(*clients.Client, *Request) error
type CloseHandler func(*clients.Client)
type HeartbeatHandler func(*clients.Client)

type authRequestHandler struct {
	rights  int32
	handler RequestHandler
}

var methods = map[string]*authRequestHandler{}
var closecallback = []CloseHandler{}
var heartbeatcallback = []HeartbeatHandler{}

func RegisterHandler(method string, rights int32, f RequestHandler) {
	_, ok := methods[method]
	if ok {
		fmt.Println("method callback conflict:", method)
	}

	methods[method] = &authRequestHandler{
		rights:  rights,
		handler: f,
	}
}

func RegisterCloseHandler(f CloseHandler) {
	closecallback = append(closecallback, f)
}

func RegisterHeartbeatHandler(f HeartbeatHandler) {
	heartbeatcallback = append(heartbeatcallback, f)
}

func OnEmit(c *clients.Client, req *Request) error {
	m := req.Method
	method, ok := methods[m]
	if !ok {
		log.Err("%v try to call %v, not found", c, m)
		return nil
	}

	err := c.Auth(method.rights)

	if m != "ClientHeartbeat" && m !="UserSrv.Heartbeat" {
		log.Debug("%v call %v %v", c, m, err)
	}

	if err != nil {
		return err
	}

	return method.handler(c, req)
}

func OnClose(c *clients.Client) {
	for _, f := range closecallback {
		f(c)
	}
}

func OnHeartbeat(c *clients.Client) {
	for _, f := range heartbeatcallback {
		f(c)
	}
}
