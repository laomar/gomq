package config

import (
	"github.com/spf13/viper"
	"log"
	"os"
	"path"
	"path/filepath"
	"strings"
)

type Listener struct {
	Type    string
	Enable  bool
	Address string
	Port    string
	Path    string
	CACert  string
	TLSCert string
	TLSKey  string
}
type Log struct {
	Level    string
	Format   string
	MaxAge   int
	MaxSize  int
	MaxCount int
}
type Config struct {
	Env       string
	DataDir   string
	PidFile   string
	Listeners map[string]Listener
	Log
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
	viper.SetDefault("tcp.port", 1883)
	viper.SetDefault("tcp.enable", true)
	viper.SetDefault("ssl.port", 8883)
	viper.SetDefault("ws.port", 8083)
	viper.SetDefault("ws.enable", true)
	viper.SetDefault("wss.port", 8084)

	// Config file
	confdir := abs("./config")
	viper.AddConfigPath("/etc/gomq")
	viper.AddConfigPath(confdir)
	viper.AddConfigPath("./config")
	viper.SetConfigName("gomq")
	if err := viper.ReadInConfig(); err != nil {
		log.Fatal(err)
	}

	// Environment variables
	viper.SetEnvPrefix("gomq")
	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	// config value
	datadir := abs(viper.GetString("datadir"))
	if _, err := os.Stat(datadir); os.IsNotExist(err) {
		if err := os.MkdirAll(datadir, 0755); err != nil {
			log.Fatal(err)
		}
	}
	lns := make(map[string]Listener)
	for _, t := range []string{"tcp", "ssl", "ws", "wss"} {
		ln := Listener{
			Type:    t,
			Enable:  viper.GetBool(t + ".enable"),
			Address: viper.GetString(t + ".address"),
			Port:    viper.GetString(t + ".port"),
		}
		if t == "ws" || t == "wss" {
			ln.Path = viper.GetString(t + ".path")
		}
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
		Log: Log{
			viper.GetString("log.level"),
			viper.GetString("log.format"),
			viper.GetInt("log.maxage"),
			viper.GetInt("log.maxsize"),
			viper.GetInt("log.maxcount"),
		},
	}
}
