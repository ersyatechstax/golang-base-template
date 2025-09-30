package nsq

import (
	"log"

	"github.com/bitly/go-nsq"

	"github.com/golang-base-template/util/config"
)

var Prods map[string]*nsq.Producer
var Nsqd *Nsq
var NsqProducersCfg map[string]*config.NSQConfig

type Nsq struct {
	producers map[string]NsqProducer
}

type NsqProducer interface {
	Publish(topic string, body []byte) error
}

func Init(appConfig config.Config) {
	NsqProducersCfg = make(map[string]*config.NSQConfig)
	for label, host := range appConfig.Nsq {
		if label != "lookupd" {
			NsqProducersCfg[label] = host
		}
	}

	Prods = make(map[string]*nsq.Producer)
	Nsqd, _ = NewNsq(nil)
	nsqConf := nsq.NewConfig()

	for key, value := range NsqProducersCfg {
		nsqProd, err := nsq.NewProducer(value.Host, nsqConf)
		if err != nil {
			log.Println("[Util][NSQ][Init] fail to create nsq producer:", err)
		}
		Prods[key] = nsqProd
		Nsqd.producers[key] = nsqProd
	}

	for label, nsqConn := range Prods {
		err := AddPublisher(label, nsqConn)
		if err != nil {
			log.Println("[Util][NSQ][Init] fail to add publisher:", err)
		}
	}

	if errMap := Ping(); len(errMap) > 0 {
		for label, err := range errMap {
			log.Printf("[Util][NSQ][Init] fail to ping nsq [%v] with error %v", label, err.Error())
		}
	}
}

func NewNsq(cfg map[string]*struct{ Host string }) (n *Nsq, err error) {
	n = &Nsq{}
	n.producers = make(map[string]NsqProducer)
	nsqconf := nsq.NewConfig()

	for key, value := range cfg {
		if nsqProd, errNsq := nsq.NewProducer(value.Host, nsqconf); errNsq == nil {
			n.producers[key] = nsqProd
		} else {
			log.Println("[util][nsqconn] fail to create nsq producer:", errNsq)
			err = errNsq
		}
	}
	return
}
