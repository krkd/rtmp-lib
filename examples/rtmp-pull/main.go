package main

import (
	"fmt"
	"net/url"
	"sync"

	"github.com/krkd/rtmp-lib/av"

	rtmp "github.com/krkd/rtmp-lib"
	"github.com/krkd/rtmp-lib/pubsub"
)

type Channel struct {
	que *pubsub.Queue
}

var channels = map[string]*Channel{}

func pullStream(url *url.URL) {

	fmt.Println(url.String())

	conn, _ := rtmp.Dial(url.String())

	_, err := conn.Streams()

	if err != nil {
		panic("got streams")
	}

	for {
		packet, err := conn.ReadPacket()
		if err != nil {
			break
		}
		fmt.Println("got packet", packet.IsKeyFrame)
	}

}

func main() {

	l := &sync.RWMutex{}

	server := &rtmp.Server{}

	server.HandlePlay = func(conn *rtmp.Conn) {

		l.RLock()
		ch := channels[conn.URL.Path]
		l.RUnlock()

		if ch != nil {

			cursor := ch.que.Latest()

			streams, err := cursor.Streams()

			if err != nil {
				panic(err)
			}

			conn.WriteHeader(streams)

			for {
				packet, err := cursor.ReadPacket()
				if err != nil {
					break
				}
				conn.WritePacket(packet)
			}
		}
	}

	server.HandlePublish = func(conn *rtmp.Conn) {

		l.Lock()
		ch := channels[conn.URL.Path]

		if ch == nil {
			ch = &Channel{}
			ch.que = pubsub.NewQueue()
			ch.que.SetMaxGopCount(1)
			channels[conn.URL.Path] = ch
		}
		l.Unlock()

		var streams []av.CodecData
		var err error

		if streams, err = conn.Streams(); err != nil {
			panic(err)
		}

		ch.que.WriteHeader(streams)

		go pullStream(conn.URL)

		for {
			var pkt av.Packet
			if pkt, err = conn.ReadPacket(); err != nil {
				break
			}

			ch.que.WritePacket(pkt)
		}

		l.Lock()
		delete(channels, conn.URL.Path)
		l.Unlock()

		ch.que.Close()

	}

	server.ListenAndServe()

}
