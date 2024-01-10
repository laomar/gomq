package config

import (
	"github.com/spf13/viper"
	. "log"
	"os"
	"path"
	"path/filepath"
	"strings"
)

type listener struct {
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
type log struct {
	Level    string
	Format   string
	MaxAge   int
	MaxSize  int
	MaxCount int
}
type mqtt struct {
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
type store struct {
	Type string
}
type Config struct {
	Env       string
	NodeName  string
	DataDir   string
	PidFile   string
	Store     store
	Listeners map[string]listener
	Log       log
	Mqtt      mqtt
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
		Fatal(err)
	}

	// config value
	datadir := abs(viper.GetString("datadir"))
	if _, err := os.Stat(datadir); os.IsNotExist(err) {
		if err := os.MkdirAll(datadir, 0755); err != nil {
			Fatal(err)
		}
	}
	lns := make(map[string]listener)
	for _, t := range []string{"tcp", "ssl", "ws", "wss"} {
		ln := listener{}
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
		PidFile:   datadir + "/" + filepath.Base(os.Args[0]) + ".pid",
		Listeners: lns,
	}
	_ = viper.UnmarshalKey("store", &Cfg.Store)
	_ = viper.UnmarshalKey("log", &Cfg.Log)
	_ = viper.UnmarshalKey("mqtt", &Cfg.Mqtt)
}
