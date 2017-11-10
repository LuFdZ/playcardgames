package clients

import (
	"fmt"
	cacheroom "playcards/model/room/cache"
	mdu "playcards/model/user/mod"
	"playcards/service/web/enum"
	"playcards/utils/auth"
	"playcards/utils/log"
	"time"
	//"runtime/debug"
	"golang.org/x/net/websocket"
)

type Message struct {
	Type string
	Data interface{}
}

type Client struct {
	user    *mdu.User
	ws      *websocket.Conn
	token   string
	topics  map[string]bool
	channel chan interface{}
	stop    chan bool
}

func NewClient(token string, u *mdu.User, ws *websocket.Conn) *Client {
	c := &Client{
		user:    u,
		ws:      ws,
		token:   token,
		topics:  make(map[string]bool),
		stop:    make(chan bool),
		channel: make(chan interface{}, 128),
	}

	Lock()
	defer Unlock()


	cs, ok := clients[u.UserID]
	if !ok {
		cs = make(map[*Client]bool)
		clients[u.UserID] = cs
	}
	//else{
	//	for oldClient,_ := range cs{
	//		//oldClient.UnsubscribeAll()
	//		//oldClient.ws.Close()
	//		DeleteClient(oldClient)
	//		close(oldClient.stop)
	//		close(oldClient.channel)
	//		cs[oldClient] = false
	//	}
	//	cs = make(map[*Client]bool)
	//	clients[u.UserID] = cs
	//	log.Debug("NewClientCoverOld:%d\n",u.UserID)
	//	//fmt.Printf("NewClientCoverOld:%d\n",u.UserID)
	//}
	//log.Debug("add connection: %v", c)
	//cs[c] = true
	//str := ""
	//for k, v := range clients {
	//	str = fmt.Sprintf("--userid:%s |\n", k)
	//	for k2, v2 := range v {
	//		str += fmt.Sprintf("----Token:%s|Topics:%#v|Ws|V:%t\n", k2.token, k2.topics,  v2)
	//	}
	//}
	//log.Debug("NewClientMap %s\n%s\n",str,string(debug.Stack()))
	//fmt.Printf("NewClientMap %s",str)
	return c
}

func (c *Client) String() string {
	return fmt.Sprintf("%s[%p]", c.user.String(), c)
}

func (c *Client) User() *mdu.User {
	return c.user
}

func (c *Client) UserID() int32 {
	return c.user.UserID
}

func (c *Client) Token() string {
	return c.token
}

func (c *Client) Auth(sright int32) error {
	return auth.Check(c.user.Rights, sright)
}

func (c *Client) Close() {
	log.Info("server close ws connection: %v", c)
	c.ws.Close()
	//断线后更新用户缓存连接状态
	cacheroom.UpdateRoomUserSocektStatus(c.User().UserID, enum.SocketClose, 0)
}

func (c *Client) Subscribe(tpc []string) {
	Lock()
	defer Unlock()
	str := fmt.Sprintf("UserSubscribeTopic:%d  :",c.UserID())
	for _, t := range tpc {
		cs, ok := topics[t]
		if !ok {
			continue
		}
		//若订阅时发现有相同userid的连接对象，则新连接覆盖老连接
		for cli,_ :=range cs{
			if cli.UserID() == c.UserID(){
				delete(cli.topics, t)
				delete(cs, cli)
				log.Debug(str+"NewSubscribeAnddeleteOld user:%v,tpc:%+v \n",cli.user,t)
			}
		}
		c.topics[t] = true
		cs[c] = true
		str += t+"|"
	}
	log.Debug(str+"\n")
}

func (c *Client) Unsubscribe(tpc []string) {
	Lock()
	defer Unlock()
	for _, tp := range tpc {
		_, ok := c.topics[tp]
		if !ok {
			return
		}

		cs := topics[tp]

		delete(c.topics, tp)
		delete(cs, c)
	}
}

// cancel all subscribe topic
func (c *Client) UnsubscribeAll() {
	Lock()
	defer Unlock()
	for t := range c.topics {
		cs, _ := topics[t]
		delete(cs, c)
	}
	c.topics = make(map[string]bool)
}

func (c *Client) SendMessage(topic, typ string, msg interface{}) {
	m := &Message{Type: typ, Data: msg}
	select {
	case c.channel <- m:
	default:
		log.Warn("drop message: %v, %v", c, m)
	}
}

func (c *Client) SendNewClientBackMessage() {
	m := &Message{Type: "SubscribeSuccess"}
	select {
	case c.channel <- m:
	default:
		log.Warn("drop message: %v, %v", c, m)
	}
}

func (c *Client) SendHearbeatMessage() {
	//m := &Message{"h"}
	select {
	case c.channel <- 1:
	default:
		log.Warn("drop hearbeat message: %v", c)
	}
}

func (c *Client) ReadLoop(f func([]byte) error, onclose func(c *Client)) {
	waitGroup.Add(1)
	defer func() {
		DeleteClient(c)
		close(c.stop)
		close(c.channel)
		onclose(c)
		waitGroup.Done()
		log.Debug("exit readloop %v", c)
	}()

	for {
		var msg []byte
		if err := websocket.Message.Receive(c.ws, &msg); err != nil {
			log.Debug("websocket recv failed c: %v msg: %v err: %v",
				c, msg, err)
			return
		}

		if err := f(msg); err != nil {
			log.Err("client process error: %v, %v", c, err)
			return
		}
	}
}

func (c *Client) Loop(heartbeat func(c *Client)) {
	waitGroup.Add(1)
	timeout := enum.HeartbeatTimeout * time.Second
	timer := time.NewTimer(timeout)

	defer func() {
		log.Debug("exit loop %v", c)
		waitGroup.Done()
		timer.Stop()
		c.Close()
	}()

	// heartbeat first
	heartbeat(c)

	for {
		select {
		case msg := <-c.channel:
			//log.Debug("websocket send: %v, %v", msg, c)
			if err := websocket.JSON.Send(c.ws, msg); err != nil {
				log.Err("websocket send error: %v, %v", msg, err)
				return
			}

		case <-c.stop:
			log.Debug("client disconnect: %v", c)
			return

		case <-timer.C:
			heartbeat(c)
			timer.Reset(timeout)
		}
	}
}
