package nsq

import (
	"time"

	"github.com/pkg/errors"
	nsq "gopkg.in/bitly/go-nsq.v1"
)

var (
	publishers map[string]Publisher
	nsqConfig  *nsq.Config
)

func init() {
	publishers = make(map[string]Publisher)
	nsqConfig = nsq.NewConfig()
}

type Publisher interface {
	Publish(topic string, data []byte) error
	DeferredPublish(topic string, delay time.Duration, data []byte) error
	Ping() error
}

func AddHost(connection, hostAddress string) (err error) {
	nsqpub, err := nsq.NewProducer(hostAddress, nsqConfig)
	if err != nil {
		return
	}
	publishers[connection] = nsqpub
	return
}

func AddPublisher(connection string, publisher Publisher) (err error) {
	if publishers == nil {
		return errors.New("failed to add publisher")
	}
	publishers[connection] = publisher
	return
}

func Connect(config map[string]*struct{ Host string }) (err error) {
	for key, value := range config {
		err = AddHost(key, value.Host)
	}

	return
}

func SendNSQ(topic string, message []byte, deferTime ...time.Duration) (err error) {
	return SendNSQTo(topic, message, "gbt", deferTime...)
}

func SendNSQTo(topic string, message []byte, connection string, deferTime ...time.Duration) (err error) {
	v, ok := publishers[connection]
	if !ok {
		return errors.Errorf("nsq connection %v not found", connection)
	}

	// if defer time is specified then defer publish
	if len(deferTime) > 0 && deferTime[0] != 0 {
		return v.DeferredPublish(topic, deferTime[0], message)
	}

	return v.Publish(topic, message)
}

func Ping() (errMap map[string]error) {
	errMap = make(map[string]error)
	for label, p := range publishers {
		if pErr := p.Ping(); pErr != nil {
			errMap[label] = pErr
		}
	}

	return errMap
}
