package server

import (
	"fmt"
	"gomq/packets"
	"net"
)

type client struct {
	conn net.Conn
}

func NewClient(conn net.Conn) *client {
	c := &client{
		conn: conn,
	}
	return c
}

func (c *client) ReadLoop() {
	for {
		_, err := packets.ReadPacket(c.conn)
		if err != nil {
			fmt.Println(err)
			break
		}
	}
}
