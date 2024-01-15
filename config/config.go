package config

import (
	"github.com/spf13/viper"
	"log"
	"os"
	"path"
	"path/filepath"
	"strings"
)

type listener struct {
	Enable        bool
	Host          string
	Port          string
	Addr          string
	ProxyProtocol bool `mapstructure:"proxy_protocol"`
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
	RetainAvailable       bool   `mapstructure:"retain_available"`
	MaxTopicAlias         uint16 `mapstructure:"max_topic_alias"`
	MaxTopicLevel         uint16 `mapstructure:"max_topic_level"`
	SessionExpiryInterval uint32 `mapstructure:"session_expiry_interval"`
	ReceiveMaximum        uint16 `mapstructure:"max_receive"`
	ServerKeepAlive       uint16 `mapstructure:"server_keep_alive"`
	MaximumPacketSize     uint32 `mapstructure:"max_packet_size"`
	MaximumQoS            byte   `mapstructure:"max_qos"`
	WildcardSub           bool   `mapstructure:"wildcard_sub"`
	SubID                 bool   `mapstructure:"sub_id"`
	SharedSub             bool   `mapstructure:"shared_sub"`
	MaxInflight           uint16 `mapstructure:"max_inflight"`
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

type config struct {
	Env       string
	DataDir   string
	PidFile   string
	NodeName  string
	Listeners map[string]*listener
	Store     store
	Mqtt      mqtt
	Log
	Plugins map[string]Config
}

type Config interface {
	Parse()
	Validate() error
}

var Cfg *config

func Default() {
	// Default value
	viper.SetDefault("env", "prod")
	viper.SetDefault("datadir", "./data")
	viper.SetDefault("tcp.port", 1883)
	viper.SetDefault("log.level", "info")

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

	// data dir
	datadir := abs(viper.GetString("datadir"))
	if _, err := os.Stat(datadir); os.IsNotExist(err) {
		if err := os.MkdirAll(datadir, 0755); err != nil {
			log.Fatal(err)
		}
	}

	// cluster dir
	nodeName := viper.GetString("cluster.node_name")
	if nodeName != "" {
		datadir += "/" + nodeName
	}

	Cfg = &config{
		Env:      viper.GetString("env"),
		DataDir:  datadir,
		PidFile:  datadir + "/gomq.pid",
		NodeName: nodeName,
		Listeners: map[string]*listener{
			"tcp": &listener{
				Enable: true,
				Port:   viper.GetString("tcp.port"),
				Addr:   ":" + viper.GetString("tcp.port"),
			},
		},
		Store: store{
			Type: "disk",
		},
		Log: Log{
			Level:    viper.GetString("log.level"),
			Format:   "json",
			MaxAge:   30,
			MaxSize:  128,
			MaxCount: 100,
		},
	}
}

// Parse Config
func (c *config) Parse() error {
	lns := make(map[string]*listener)
	for _, t := range []string{"tcp", "tls", "ws", "wss"} {
		ln := &listener{}
		if err := viper.UnmarshalKey(t, ln); err != nil {
			return err
		}
		if t == "tls" || t == "wss" {
			ln.CACert = abs(viper.GetString(t + ".cacert"))
			ln.TLSCert = abs(viper.GetString(t + ".tlscert"))
			ln.TLSKey = abs(viper.GetString(t + ".tlskey"))
		}
		ln.Addr = ln.Host + ":" + ln.Port
		lns[t] = ln
	}
	Cfg.Listeners = lns

	if err := viper.UnmarshalKey("log", &Cfg.Log); err != nil {
		return err
	}
	if err := viper.UnmarshalKey("store", &Cfg.Store); err != nil {
		return err
	}
	if err := viper.UnmarshalKey("mqtt", &Cfg.Mqtt); err != nil {
		return err
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
	Default()
	if err := Cfg.Parse(); err != nil {
		log.Fatal(err)
	}
}

func RegPluginConfig(name string, cfg Config) {

}
