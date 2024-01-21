package main

import (
	"fmt"

	rtmp "github.com/krkd/rtmp-lib"
)

func main() {

	server := &rtmp.Server{}

	server.HandlePublish = func(conn *rtmp.Conn) {

		for {
			packet, err := conn.ReadPacket()
			if err != nil {
				break
			}
			fmt.Println(packet.Time)
		}
	}

	server.ListenAndServe()

}
