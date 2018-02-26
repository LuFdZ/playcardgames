package clients

import (
	"errors"
	"playcards/utils/log"
	"sync"
	"fmt"
	"bytes"
	"playcards/utils/tools"
)

var waitGroup = new(sync.WaitGroup)

var clients = make(map[int32]map[*Client]bool)
var topics = make(map[string]map[*Client]bool)
var lock = new(sync.RWMutex)

func InitTopic(topic string) error {
	log.Info("service init topic: %v", topic)
	Lock()
	defer Unlock()

	_, ok := topics[topic]
	if ok {
		return errors.New("already init")
	}

	topics[topic] = make(map[*Client]bool)
	return nil
}

func Done() {
	RUlock()
}

func GetLockClients(topic string) map[*Client]bool {
	RLock()
	return topics[topic]
}

func GetClients() map[int32]map[*Client]bool {
	return clients
}

func GetClientByUserID(uid int32) []*Client {
	var out []*Client
	Lock()
	defer Unlock()

	val, ok := clients[uid]
	if !ok {
		return out
	}
	for v, _ := range val {
		out = append(out, v)
	}
	return out
}

func RLock() {
	lock.RLock()
}

func RUlock() {
	lock.RUnlock()
}

func Lock() {
	lock.Lock()
}

func Unlock() {
	lock.Unlock()
}

func SendClientMsg(uid int32, topic, typ string, msg interface{}, code int) {
	if cs, ok := clients[uid]; ok {
		for k, v := range cs {
			if v {
				k.SendMessage(topic, typ, msg, code)
			}
		}
	}
}

func Send(topic, typ string, msg interface{}, code int) error {
	return SendWhere(topic, typ, msg, nil, code)
}

func SendTo(uid int32, topic, typ string, msg interface{}, code int) error {
	return SendWhere(topic, typ, msg, func(c *Client) bool {
		return c.UserID() == uid
	}, code)
}

func SendToBackLog(uid int32, topic, typ string, msg interface{}, b *bytes.Buffer, code int) error {
	cs := GetLockClients(topic)
	defer Done()
	for c, _ := range cs {
		if c.UserID() != uid {
			continue
		}
		b.WriteString(tools.IntToString(c.UserID()))
		b.WriteString("|")
		ws := fmt.Sprintf("%v", &c.ws)
		b.WriteString(ws)
		b.WriteString(",")

		c.SendMessage(topic, typ, msg, code)
	}
	return nil
}

func SendWhere(topic, typ string, msg interface{},
	f func(*Client) bool, code int) error {
	cs := GetLockClients(topic)
	defer Done()
	str := fmt.Sprintf("sendwhere:%s,", topic)
	for c, _ := range cs {
		if f != nil && !f(c) {
			continue
		}
		str += fmt.Sprintf("### %d|%v ###", c.UserID(), &c.ws)
		c.SendMessage(topic, typ, msg, code)
	}
	log.Debug(str)
	return nil
}

func SendToUsers(ids []int32, topic, typ string, msg interface{}, code int) error {
	b := &bytes.Buffer{}

	for _, id := range ids {
		SendToBackLog(id, topic, typ, msg, b, code)
	}
	log.Debug("send to users sent:%v,success:%s,typ:%s,msg:%v", ids, b.String(), typ, msg)
	return nil
}

func DeleteClient(c *Client) {
	log.Debug("delete connection: %v", c)

	Lock()
	defer Unlock()

	uid := c.UserID()
	cs := clients[uid]

	delete(cs, c)
	if len(cs) == 0 {
		delete(clients, uid)
	}

	for t, _ := range c.topics {
		delete(topics[t], c)
	}
}

func CloseAll() {
	Lock()
	for _, cs := range clients {
		if len(cs) == 0 {
			break
		}
		for c, _ := range cs {
			c.Close()
		}
	}
	Unlock()
	waitGroup.Wait()
}

//func PageClients(page int32,size int32,total int32) []int32{
//
//}
