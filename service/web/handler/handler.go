package handler

import (
	"encoding/json"
	pbweb "playcards/proto/web"
	"playcards/service/web/clients"
	"playcards/service/web/publish"
	"playcards/service/web/request"
	"playcards/utils/auth"
	"playcards/utils/log"

	"github.com/micro/go-micro/broker"
	"github.com/micro/go-micro/client"
	"golang.org/x/net/context"
	"golang.org/x/net/websocket"
)

type Web struct {
	client client.Client
	broker broker.Broker
}

func NewWebHandler(c client.Client) *Web {
	// make client's broker connect to cluster
	topic := publish.TopicServiceOnline
	msg := c.NewPublication(topic, &pbweb.ServiceOnline{})
	c.Publish(context.Background(), msg)

	w := &Web{
		client: c,
		broker: c.Options().Broker,
	}

	return w
}

func (w *Web) Subscribe(ws *websocket.Conn) {
	var msg []byte
	if err := websocket.Message.Receive(ws, &msg); err != nil {
		log.Err("websocket recv error: %v", err)
		return
	}

	token := string(msg)
	u, err := auth.GetUserByToken(token)
	if err != nil {
		log.Err("websocket login failed: %v token: %v", err, token)
		return
	}

	c := clients.NewClient(token, u, ws)
	log.Debug("new client: %v", c)

	f := func(msg []byte) error {
		req := &request.Request{}
		err := json.Unmarshal(msg, &req)
		if err != nil {
			log.Err("client %v request error: %v", c, req)
			return err
		}
		return request.OnEmit(c, req)
	}

	go c.ReadLoop(f, request.OnClose)
	c.Loop(request.OnHeartbeat)

	log.Debug("%v stream broken", c)
}

func UnsubscribeAll(c *clients.Client, req *request.Request) error {
	c.UnsubscribeAll()
	return nil
}

func init() {
	request.RegisterHandler("UnsubscribeAll", auth.RightsPlayer,
		UnsubscribeAll)
}

func (w *Web) Stop() {
	clients.CloseAll()
}
