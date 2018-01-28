package clients

import (
	"fmt"
	mdu "playcards/model/user/mod"
	"playcards/service/web/enum"
	"playcards/utils/auth"
	"playcards/utils/log"
	"time"
	"golang.org/x/net/websocket"
	"bytes"
)

type Message struct {
	Type string
	Data interface{}
	Cmd  int
	WsID string
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
	} else {
		for k, _ := range cs {
			k.Close()
		}
	}
	log.Debug("connection count:%d, add connection: %v", len(cs), c)
	cs[c] = true
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
}

func (c *Client) Subscribe(tpc []string) {
	Lock()
	defer Unlock()
	b := &bytes.Buffer{}
	for _, t := range tpc {
		cs, ok := topics[t]
		if !ok {
			continue
		}
		c.topics[t] = true
		cs[c] = true
		b.WriteString(t)
		b.WriteString("|")
		//log.Debug("%v subscribe topic [%v]", c, t)
	}
	log.Debug("%v subscribe topic [%s]", c, b.String())
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

func (c *Client) SendMessage(topic, typ string, msg interface{}, code int) {
	m := &Message{Type: typ, Data: msg, WsID: fmt.Sprintf("%v", &c.ws), Cmd: code}
	select {
	case c.channel <- m:
	default:
		log.Warn("drop message: %v, %v", c, m)
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
			log.Err("client process error: %v, %v,%v", c, err, string(msg))
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
