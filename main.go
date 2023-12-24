package main

import (
	"fmt"
	. "github.com/laomar/gomq/config"
	"github.com/laomar/gomq/log"
	. "github.com/laomar/gomq/server"
	"github.com/spf13/cobra"
	"github.com/syndtr/goleveldb/leveldb"
)

// Root command
var rootCmd = &cobra.Command{
	Use:     "gomqd",
	Short:   "Gomq is a high-performance MQTT broker for IoT",
	Version: "1.0.0",
}

func init() {
	ParseConfig()
	log.Init()
}

func main() {
	db, _ := leveldb.OpenFile(Cfg.DataDir+"/session", nil)
	defer db.Close()
	_ = db.Put([]byte("session"), []byte("test"), nil)
	data, _ := db.Get([]byte("session"), nil)
	fmt.Println(string(data))
	rootCmd.AddCommand(StartCmd(), StopCmd(), ReloadCmd())
	if err := rootCmd.Execute(); err != nil {
		log.Errorf("Cmd: %v", err)
	}
}
