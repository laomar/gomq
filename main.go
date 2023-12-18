package main

import (
	"github.com/spf13/cobra"
	"gomq/config"
	"gomq/log"
	"gomq/server"
)

var rootCmd = &cobra.Command{
	Use:     "gomqd",
	Short:   "Gomq is a high-performance MQTT broker for IoT",
	Version: "1.0.0",
}

var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Start gomq broker",
	Run: func(cmd *cobra.Command, args []string) {
		server := server.New()
		server.Start()
	},
}

func init() {
	config.Parse()
	log.Init()
}

func main() {
	rootCmd.AddCommand(startCmd)
	if err := rootCmd.Execute(); err != nil {
		log.Fatal(err)
	}
}
