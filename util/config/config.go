package config

import (
	"flag"
	"io"
	"log"
	"os"
	"path/filepath"
	"sync"

	"github.com/bitly/go-nsq"
	"github.com/pkg/errors"
	"gopkg.in/yaml.v3"
)

var (
	cfg      Config
	mux      sync.Mutex
	hostname string

	isUnitTest = false
)

type (
	Config struct {
		ServiceName  string
		Port         PortConfig
		Database     map[string]*DatabaseConf `json:"database"`
		Redis        map[string]*RedisConf    `json:"redis"`
		Nsq          map[string]*NSQConfig    `yaml:"nsq"`
		Consumer     ConsumerConfig
		ConsumerList map[string]*ConsumerListConfig `yaml:"ConsumerList"`
		URL          URLConfig
	}

	// PortConfig is config for app port
	PortConfig struct {
		GBT  string
		Grpc string
		Bg   string
	}

	// DatabaseConf is config for database
	DatabaseConf struct {
		MasterMaxConn int    `gcfg:"master-conn"`
		MasterMaxIdle int    `gcfg:"master-idle"`
		SlaveMaxConn  int    `gcfg:"slave-conn"`
		SlaveMaxIdle  int    `gcfg:"slave-idle"`
		DbName        string `gcfg:"dbname"`
		MasterHost    string `gcfg:"master-host"`
		SlaveHost     string `gcfg:"slave-host"`
		PortMaster    string `gcfg:"port-master"`
		PortSlave     string `gcfg:"port-slave"`
		User          string `gcfg:"user"`
		Password      string `gcfg:"password"`
	}
	// RedisConf is config for redis
	RedisConf struct {
		Address          string `gcfg:"address"`
		PoolSize         int    `gcfg:"pool-size"`
		PoolTimeoutMS    int    `gcfg:"pool-timeout"`
		DialTimeoutMS    int    `gcfg:"dial-timeout"`
		ReadTimeoutMS    int    `gcfg:"read-timeout"`
		WriteTimeoutMS   int    `gcfg:"write-timeout"`
		IdleTimeoutSec   int    `gcfg:"idle-timeout-sec"`
		IdleFreqCheckSec int    `gcfg:"idle-frequency-check-sec"`
	}
	//ConsumerConfig contains default configuration for nsq consumers
	ConsumerConfig struct {
		LookupdAddress      []string
		DefaultMaxInflight  int
		DefaultMaxAttempts  uint16
		MaxBackoffDuration  int
		DefaultRequeueDelay int
	}

	//ConsumerListConfig shall only be used by consumer package
	ConsumerListConfig struct {
		Switch       bool   `yaml:"Switch"`
		Topic        string `yaml:"Topic"`
		Channel      string `yaml:"Channel"`
		WorkerAmount int    `yaml:"WorkerAmount"`
		MaxInFlight  int    `yaml:"MaxInFlight"`

		Handler nsq.Handler
		Config  *nsq.Config
	}

	//NSQConfig for NSQ publisher config
	NSQConfig struct {
		Host string `yaml:"host"`
	}

	URLConfig struct {
		ExampleExternalService string
	}
)

func InitConfig() error {
	// if run from unit test
	if flag.Lookup("test.v") != nil {
		if flag.Lookup("test.v").Value.String() == "true" {
			cfg = Config{
				Port: PortConfig{
					GBT:  "13001",
					Bg:   "13002",
					Grpc: "13003",
				},
			}
			return nil
		}
	}

	environ := os.Getenv("ENV")
	if environ == "" {
		environ = "development"
	}

	err := Read(&cfg, "configs/etc/gbt/gbt."+environ+".yaml")
	if err != nil {
		log.Printf("Error read from configs/etc/gbt folder, err: %v\n", err)
		err := Read(&cfg, "/etc/gbt/gbt."+environ+".yaml")
		if err != nil {
			log.Printf("Error read from /etc/gbt folder, err: %v\n", err)
			return errors.Wrap(err, "failed to read config")
		}
	}

	return nil
}

// Get is for get curent config value
func Get() Config {
	mux.Lock()
	defer mux.Unlock()
	// if config is empty, re init
	// port is never empty
	if cfg.Port.GBT == "" || cfg.Port.Bg == "" || cfg.Port.Grpc == "" {
		InitConfig()
	}

	return cfg
}

// Set is for set new config replacing current config
func Set(config Config) {
	mux.Lock()
	defer mux.Unlock()
	cfg = config
}

func Read(dest interface{}, paths ...string) error {
	for _, path := range paths {
		// check if path exists
		if _, err := os.Stat(path); os.IsNotExist(err) {
			continue
		}

		// load config
		ext := filepath.Ext(path)
		f, err := os.Open(path)
		if err != nil {
			return err
		}
		content, err := io.ReadAll(f)
		if err != nil {
			return err
		}

		switch ext {
		case ".yml", ".yaml":
			return yaml.Unmarshal(content, dest)
		}
	}

	return errors.New("config file not found")
}
