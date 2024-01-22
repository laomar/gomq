package config

import (
	"github.com/mitchellh/mapstructure"
	"github.com/spf13/viper"
	"log"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

type listener struct {
	Enable        bool
	Host          string
	Port          int
	Addr          string
	ProxyProtocol bool `toml:"proxy_protocol"`
	Path          string
	CACert        string
	TLSCert       string
	TLSKey        string
}

type Log struct {
	Level    string
	Format   string
	MaxAge   int
	MaxSize  int
	MaxCount int
}

type mqtt struct {
	RetainAvailable       bool   `toml:"retain_available"`
	MaxTopicAlias         uint16 `toml:"max_topic_alias"`
	MaxTopicLevel         uint16 `toml:"max_topic_level"`
	SessionExpiryInterval uint32 `toml:"session_expiry_interval"`
	ReceiveMaximum        uint16 `toml:"max_receive"`
	ServerKeepAlive       uint16 `toml:"server_keep_alive"`
	MaximumPacketSize     uint32 `toml:"max_packet_size"`
	MaximumQoS            byte   `toml:"max_qos"`
	WildcardSub           bool   `toml:"wildcard_sub"`
	SubID                 bool   `toml:"sub_id"`
	SharedSub             bool   `toml:"shared_sub"`
	MaxInflight           uint16 `toml:"max_inflight"`
}

type store struct {
	Type  string
	Redis redis
}

type redis struct {
	Addrs []string
	User  string
	Pwd   string
}

type cluster struct {
	NodeName         string        `toml:"node_name"`
	GrpcPort         int           `toml:"grpc_port"`
	GossipHost       string        `toml:"gossip_host"`
	GossipPort       int           `toml:"gossip_port"`
	RetryJoin        []string      `toml:"retry_join"`
	RetryInterval    time.Duration `toml:"retry_interval"`
	RetryTimeout     time.Duration `toml:"retry_timeout"`
	RejoinAfterLeave bool          `toml:"rejoin_after_leave"`
}

type config struct {
	Env       string
	DataDir   string
	PidFile   string
	NodeName  string
	Listeners map[string]*listener
	Store     store
	Mqtt      mqtt
	Cluster   cluster
	Log       Log
	Plugins   map[string]Config
}

type Config interface {
	Validate() error
}

var Cfg *config
var DecoderConfigOption = func(c *mapstructure.DecoderConfig) {
	c.TagName = "toml"
}

// Init Config
func Init() {
	// Default value
	viper.SetDefault("env", "prod")
	viper.SetDefault("datadir", "./data")
	viper.SetDefault("tcp.port", 1883)
	viper.SetDefault("log.level", "info")
	viper.SetDefault("grpc.port", 8866)
	viper.SetDefault("gossip.port", 8666)

	// Environment variables
	viper.SetEnvPrefix("gomq")
	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	// Config file
	conf := viper.GetString("conf")
	if conf != "" {
		viper.SetConfigFile(conf)
	} else {
		viper.AddConfigPath("/etc/gomq")
		viper.AddConfigPath(abs("./config"))
		viper.AddConfigPath("./config")
		viper.SetConfigName("gomq")
		viper.SetConfigType("toml")
	}
	if err := viper.ReadInConfig(); err != nil {
		log.Fatal(err)
	}

	Cfg = &config{
		Env:      viper.GetString("env"),
		DataDir:  viper.GetString("datadir"),
		NodeName: viper.GetString("cluster.node_name"),
		Listeners: map[string]*listener{
			"tcp": &listener{
				Enable: true,
				Port:   viper.GetInt("tcp.port"),
				Addr:   ":" + viper.GetString("tcp.port"),
			},
		},
		Store: store{
			Type: "disk",
		},
		Cluster: cluster{
			NodeName:         viper.GetString("cluster.node_name"),
			GrpcPort:         viper.GetInt("grpc.port"),
			GossipPort:       viper.GetInt("gossip.port"),
			RetryInterval:    5 * time.Second,
			RetryTimeout:     30 * time.Second,
			RejoinAfterLeave: true,
		},
		Mqtt: mqtt{
			RetainAvailable:       true,
			MaxTopicAlias:         65535,
			MaxTopicLevel:         128,
			SessionExpiryInterval: 60,
			ReceiveMaximum:        128,
			ServerKeepAlive:       0,
			MaximumPacketSize:     10240,
			MaximumQoS:            2,
			WildcardSub:           true,
			SubID:                 true,
			SharedSub:             true,
			MaxInflight:           32,
		},
		Log: Log{
			Level:    viper.GetString("log.level"),
			Format:   "json",
			MaxAge:   30,
			MaxSize:  128,
			MaxCount: 100,
		},
		Plugins: make(map[string]Config),
	}
}

// Parse Config
func (c *config) Parse() error {
	// Parse listener
	lns := make(map[string]*listener)
	for _, t := range []string{"tcp", "tls", "ws", "wss"} {
		ln := &listener{}
		if err := viper.UnmarshalKey(t, ln, DecoderConfigOption); err != nil {
			return err
		}
		if t == "tls" || t == "wss" {
			ln.CACert = abs(ln.CACert)
			ln.TLSCert = abs(ln.TLSCert)
			ln.TLSKey = abs(ln.TLSKey)
		}
		ln.Addr = ln.Host + ":" + strconv.Itoa(ln.Port)
		lns[t] = ln
	}
	Cfg.Listeners = lns

	if err := viper.Unmarshal(Cfg, DecoderConfigOption); err != nil {
		return err
	}

	// Parse data dir
	datadir := abs(Cfg.DataDir)
	if _, err := os.Stat(datadir); os.IsNotExist(err) {
		if err := os.MkdirAll(datadir, 0755); err != nil {
			return err
		}
	}
	Cfg.DataDir = datadir
	Cfg.PidFile = datadir + "/gomq.pid"

	// Parse plugin
	for name, cfg := range c.Plugins {
		if err := viper.UnmarshalKey(name, cfg, DecoderConfigOption); err != nil {
			return err
		}
	}

	return c.Validate()
}

// Validate Config
func (c *config) Validate() error {
	return nil
}

func abs(p string) string {
	if strings.HasPrefix(p, "/") {
		return p
	}
	appdir, _ := filepath.Abs(filepath.Dir(os.Args[0]))
	return path.Join(appdir, p)
}

func init() {
	Init()
	if err := Cfg.Parse(); err != nil {
		log.Fatal(err)
	}
}

func ParsePlugin(name string, cfg Config) {
	if err := viper.UnmarshalKey(name, cfg, DecoderConfigOption); err != nil {
		log.Fatal(err)
	}
	Cfg.Plugins[name] = cfg
}
