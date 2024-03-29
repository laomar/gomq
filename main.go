package main

import (
	"github.com/laomar/gomq/log"
	. "github.com/laomar/gomq/server"
	"github.com/spf13/cobra"
)

// Root command
var rootCmd = &cobra.Command{
	Use:     "gomqd",
	Short:   "Gomq is a high-performance MQTT broker for IoT",
	Version: "0.1.0",
}

func main() {
	rootCmd.AddCommand(StartCmd(), StopCmd(), ReloadCmd())
	if err := rootCmd.Execute(); err != nil {
		log.Errorf("Cmd: %v", err)
	}
}
