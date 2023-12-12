package main

import (
	"gomq/server"
)

func main() {
	s := server.New()
	s.Run()
}
