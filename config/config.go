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
	Listeners map[string]Listener
	Log
}

var Cfg *Config

func Parse() {
	// default value
	viper.SetDefault("env", "prod")
	viper.SetDefault("datadir", "./data")
	viper.SetDefault("tcp.port", 1883)
	viper.SetDefault("tcp.enable", true)
	viper.SetDefault("ssl.port", 8883)
	viper.SetDefault("ws.port", 8083)
	viper.SetDefault("ws.enable", true)
	viper.SetDefault("wss.port", 8084)
	// config file
	viper.AddConfigPath("./config")
	viper.AddConfigPath("/etc/gomq")
	viper.SetConfigName("gomq")
	if err := viper.ReadInConfig(); err != nil {
		log.Println(err)
	}
	// environment variables
	viper.SetEnvPrefix("gomq")
	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	// config value
	datadir := viper.GetString("datadir")
	if !strings.HasPrefix(datadir, "/") {
		datadir = path.Join(filepath.Dir(os.Args[0]), datadir)
	}
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
			ln.CACert = viper.GetString(t + ".cacert")
			ln.TLSCert = viper.GetString(t + ".tlscert")
			ln.TLSKey = viper.GetString(t + ".tlskey")
		}
		lns[t] = ln
	}
	Cfg = &Config{
		Env:       viper.GetString("env"),
		DataDir:   datadir,
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
