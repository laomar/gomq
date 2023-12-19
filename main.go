package main

import (
	"github.com/spf13/cobra"
	. "gomq/config"
	"gomq/log"
	. "gomq/server"
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
	rootCmd.AddCommand(StartCmd(), StopCmd(), ReloadCmd())
	if err := rootCmd.Execute(); err != nil {
		log.Errorf("Cmd: %v", err)
	}
}
