package cluster

import (
	"fmt"
	"github.com/hashicorp/serf/serf"
	"github.com/laomar/gomq/config"
	"github.com/laomar/gomq/log"
	"github.com/laomar/gomq/server"
	"go.uber.org/zap"
)

const Name = "cluster"

func init() {
	server.RegPlugin(Name, New)
	config.RegPluginConfig(Name, Cfg)
}

type cluster struct {
	nodeName    string
	serf        *serf.Serf
	serfEventCh chan serf.Event
}

func New() (server.Plugin, error) {
	fmt.Println(Cfg.NodeName)
	c := &cluster{
		nodeName:    "gomq",
		serfEventCh: make(chan serf.Event, 1000),
	}

	logger := zap.NewStdLog(log.Logger)
	sf := serf.DefaultConfig()
	sf.Logger = logger
	sf.MemberlistConfig.BindPort = 8421
	sf.MemberlistConfig.AdvertiseAddr = "0.0.0.0"
	sf.MemberlistConfig.AdvertisePort = 8422
	sf.MemberlistConfig.Logger = logger
	sf.NodeName = "gomq"
	//s, err := serf.Create(sf)
	//if err != nil {
	//	return nil, err
	//}
	//c.serf = s

	return c, nil
}

func (c *cluster) Name() string {
	return Name
}

func (c *cluster) Load() error {
	return nil
}

func (c *cluster) Unload() error {
	if err := c.serf.Leave(); err != nil {
		return err
	}
	return c.serf.Shutdown()
}
