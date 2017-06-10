package topic

import (
	"fmt"
	"playcards/utils/log"

	"github.com/micro/go-micro/broker"
	"github.com/micro/protobuf/proto"
)

func Topic(t string, args ...interface{}) string {
	return fmt.Sprintf(t, args...)
}

func Publish(brk broker.Broker, msg proto.Message, t string,
	args ...interface{}) error {

	topic := Topic(t, args...)
	log.Debug("pub %v: %v", topic, msg)
	body, err := proto.Marshal(msg)
	if err != nil {
		log.Warn("pub err: %v %v, %v", topic, msg, err)
		return err
	}
	m := &broker.Message{Body: body}

	return brk.Publish(topic, m)
}
