package config

import (
	"fmt"
	"github.com/spf13/viper"
	"log"
	"os"
	"path"
	"path/filepath"
	"strings"
)

type Listener struct {
	Type          string
	Enable        bool
	Address       string
	Port          string
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
type Mqtt struct {
	RetainAvailable       bool   `mapstructure:"retain_available"`
	TopicAliasMaximum     uint16 `mapstructure:"max_topic_alias"`
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
type Config struct {
	Env       string
	DataDir   string
	PidFile   string
	Listeners map[string]Listener
	Log       Log
	Mqtt      Mqtt
}

var Cfg *Config

func abs(p string) string {
	if strings.HasPrefix(p, "/") {
		return p
	}
	appdir, _ := filepath.Abs(filepath.Dir(os.Args[0]))
	return path.Join(appdir, p)
}

func ParseConfig() {
	// Default value
	viper.SetDefault("env", "prod")
	viper.SetDefault("datadir", "./data")

	// Environment variables
	viper.SetEnvPrefix("gomq")
	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	// Config file
	confdir := abs("./config")
	viper.AddConfigPath("/etc/gomq")
	viper.AddConfigPath(confdir)
	viper.AddConfigPath("./config")
	viper.SetConfigName("gomq")
	if err := viper.ReadInConfig(); err != nil {
		log.Fatal(err)
	}

	// config value
	datadir := abs(viper.GetString("datadir"))
	if _, err := os.Stat(datadir); os.IsNotExist(err) {
		if err := os.MkdirAll(datadir, 0755); err != nil {
			log.Fatal(err)
		}
	}
	lns := make(map[string]Listener)
	for _, t := range []string{"tcp", "ssl", "ws", "wss"} {
		ln := Listener{}
		_ = viper.UnmarshalKey(t, &ln)
		if t == "ssl" || t == "wss" {
			ln.CACert = abs(viper.GetString(t + ".cacert"))
			ln.TLSCert = abs(viper.GetString(t + ".tlscert"))
			ln.TLSKey = abs(viper.GetString(t + ".tlskey"))
		}
		lns[t] = ln
	}

	Cfg = &Config{
		Env:       viper.GetString("env"),
		DataDir:   datadir,
		PidFile:   abs(viper.GetString("pidfile")),
		Listeners: lns,
	}
	_ = viper.UnmarshalKey("log", &Cfg.Log)
	_ = viper.UnmarshalKey("mqtt", &Cfg.Mqtt)

	fmt.Println(Cfg.Listeners["tcp"])
}
