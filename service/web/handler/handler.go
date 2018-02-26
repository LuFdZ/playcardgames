package handler

import (
	"encoding/json"
	pbweb "playcards/proto/web"
	webbill "playcards/service/web/bill"
	"playcards/service/web/clients"
	webclub "playcards/service/web/club"
	webniu "playcards/service/web/niuniu"
	"playcards/service/web/publish"
	"playcards/service/web/request"
	webroom "playcards/service/web/room"
	webthirteen "playcards/service/web/thirteen"
	webdoudizhu "playcards/service/web/doudizhu"
	webfour "playcards/service/web/fourcard"
	webuser "playcards/service/web/user"
	webmail "playcards/service/web/mail"
	webtow "playcards/service/web/twocard"
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
	webuser.Init(w.broker)
	webbill.Init(w.broker)
	webroom.Init(w.broker)
	webthirteen.Init(w.broker)
	webniu.Init(w.broker)
	webclub.Init(w.broker)
	webdoudizhu.Init(w.broker)
	webfour.Init(w.broker)
	webmail.Init(w.broker)
	webtow.Init(w.broker)
	return w
}

func (w *Web) Subscribe(ws *websocket.Conn) {
	var msg []byte

	if err := websocket.Message.Receive(ws, &msg); err != nil {
		log.Err("websocket recv error: %v|%s", err, string(msg))
		return
	}
	//log.Err("Subscribe websocket recv\n: %+v", string(msg))
	token := string(msg)
	u, err := auth.GetUserByToken(token)
	if err != nil {
		log.Err("websocket login failed: %v token: %v", err, token)
		return
	}
	if u == nil {
		log.Err("websocket get user null: %v", string(msg))
		return
	}
	c := clients.NewClient(token, u, ws)

	webroom.SubscribeRoomMessage(c, nil)
	webbill.SubscribeBillMessage(c, nil)
	webthirteen.SubscribeThirteenMessage(c, nil)
	webniu.SubscribeNiuniuMessage(c, nil)
	webdoudizhu.SubscribeDoudizhuMessage(c, nil)
	webfour.SubscribeFourCardMessage(c, nil)
	webmail.SubscribeMailMessage(c, nil)
	webtow.SubscribeTwoCardMessage(c, nil)

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
	webclub.ClubOnlineNotice(c)
	webroom.RoomOnlineNotice(c)
	go c.ReadLoop(f, request.OnClose)
	c.Loop(request.OnHeartbeat)

	log.Debug("%v stream broken", c)

}

func UnsubscribeAll(c *clients.Client, req *request.Request) error {
	c.UnsubscribeAll()
	return nil
}

func UserWebScoketList(c *clients.Client, req *request.Request) error {
	cs := clients.GetClients()
	count := int32(len(cs))
	var ids []int32
	for k, _ := range cs {
		ids = append(ids, k)
	}
	msg := &pbweb.UserWebScoketList{
		Count: count,
		Ids:   ids,
	}
	c.SendMessage("WebScoketCount", "WebScoketCount", msg, 0)
	return nil
}

func init() {
	request.RegisterHandler("UnsubscribeAll", auth.RightsPlayer,
		UnsubscribeAll)
	request.RegisterHandler("UserWebScoketList", auth.RightsAdmin,
		UserWebScoketList)
}

func (w *Web) Stop() {
	clients.CloseAll()

}
