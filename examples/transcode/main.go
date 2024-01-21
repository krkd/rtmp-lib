package main

import (
	"fmt"
	_ "github.com/krkd/rtmp-lib/aac"
	"github.com/krkd/rtmp-lib/audio"
	"github.com/krkd/rtmp-lib/av"
	"os"

	rtmp "github.com/krkd/rtmp-lib"
)

func main() {

	server := &rtmp.Server{}

	server.HandlePublish = func(conn *rtmp.Conn) {

		file, err := os.Create("test.opus")
		if err != nil {
			panic(err)
		}

		streams, err := conn.Streams()

		var enc av.AudioEncoder
		var dec av.AudioDecoder

		var adecodec av.AudioCodecData

		for _, stream := range streams {
			if stream.Type().IsAudio() {
				adecodec = stream.(av.AudioCodecData)
				dec, _ = audio.NewAudioDecoderByName("aac")
				dec.SetSampleRate(adecodec.SampleRate())
				dec.SetSampleFormat(av.S16)
				dec.SetChannelLayout(adecodec.ChannelLayout())
				err = dec.Setup()
				if err != nil {
					fmt.Println(err)
				}
				enc, err = audio.NewAudioEncoderByName("libopus")
				if err != nil {
					fmt.Println(err)
				}
				//enc.SetSampleFormat(av.S16)
				enc.SetSampleRate(48000)
				enc.SetSampleFormat(av.S16)
				enc.SetChannelLayout(av.CH_STEREO)
				enc.Setup()

			}
		}

		for {
			packet, err := conn.ReadPacket()
			if err != nil {
				break
			}

			stream := streams[packet.Idx]

			if stream.Type().IsVideo() {
				continue
			}

			ok, frame, err := dec.Decode(packet.Data)
			if err != nil {
				fmt.Println(err)
				continue
			}
			if !ok {
				continue
			}

			dur, _ := dec.PacketDuration(packet.Data)
			fmt.Println("decode dur", dur)

			var _outpkts [][]byte
			if _outpkts, err = enc.Encode(frame); err != nil {
				fmt.Println(err)
				continue
			}

			for _, outpkt := range _outpkts {
				//adtsbuffer := []byte{}
				//adtsheader := make([]byte, 7)
				//aac.FillADTSHeader(adtsheader, aencodec.(aac.CodecData).Config, 1024, len(outpkt))
				//adtsbuffer = append(adtsbuffer, adtsheader...)
				//adtsbuffer = append(adtsbuffer, outpkt...)
				dur, _ := enc.PacketDuration(outpkt)
				file.Write(outpkt)
				fmt.Println("encode dur", dur)
				//
				//dur, _:= aencodec.PacketDuration(outpkt)
				//fmt.Println("encode dur", outpkt)

			}

			fmt.Println(packet.Time)
		}
	}

	server.ListenAndServe()

}
