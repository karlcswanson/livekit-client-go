package main

import (
	"errors"
	"fmt"
	"log"
	"os"
	"os/signal"
	"runtime"
	"strings"
	"syscall"

	"github.com/joho/godotenv"
	gst "github.com/karlcswanson/livekit-client-go/internal/gstreamer-sink"
	"github.com/pion/rtp/codecs"
	"github.com/pion/webrtc/v3"

	lksdk "github.com/livekit/server-sdk-go"
	"github.com/livekit/server-sdk-go/pkg/samplebuilder"
)

var (
	host, apiKey, apiSecret, roomName, identity string
)

func init() {
	runtime.LockOSThread()
}

func newMain() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}
	host = os.Getenv("LIVEKIT_URL")
	apiKey = os.Getenv("LIVEKIT_API_KEY")
	apiSecret = os.Getenv("LIVEKIT_API_SECRET")
	roomName = os.Getenv("LIVEKIT_ROOM")
	identity = os.Getenv("LIVEKIT_ID")
	if host == "" || apiKey == "" || apiSecret == "" || roomName == "" || identity == "" {
		fmt.Println("invalid arguments.")
		return
	}
	room, err := lksdk.ConnectToRoom(host, lksdk.ConnectInfo{
		APIKey:              apiKey,
		APISecret:           apiSecret,
		RoomName:            roomName,
		ParticipantIdentity: identity,
	})
	if err != nil {
		panic(err)
	}
	room.Callback.OnTrackSubscribed = onTrackSubscribed
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT)

	<-sigChan
	room.Disconnect()
}

func main() {
	go newMain()
	gst.StartMainLoop()
}

func onTrackSubscribed(track *webrtc.TrackRemote, publication *lksdk.RemoteTrackPublication, rp *lksdk.RemoteParticipant) {
	fileName := fmt.Sprintf("%s-%s", rp.Identity(), track.ID())
	fmt.Println("write track to file ", fileName)
	NewTrackWriter(track, rp.WritePLI, fileName)
}

const (
	maxAudioLate = 200 // 4s for audio
)

type TrackWriter struct {
	sb       *samplebuilder.SampleBuilder
	pipeline *gst.Pipeline
	track    *webrtc.TrackRemote
}

func NewTrackWriter(track *webrtc.TrackRemote, pliWriter lksdk.PLIWriter, fileName string) (*TrackWriter, error) {
	var (
		sb       *samplebuilder.SampleBuilder
		pipeline *gst.Pipeline
		err      error
	)
	switch {
	case strings.EqualFold(track.Codec().MimeType, "audio/opus"):
		sb = samplebuilder.New(maxAudioLate, &codecs.OpusPacket{}, track.Codec().ClockRate)
		pipeline = gst.CreatePipeline(track.PayloadType(), "opus")

	default:
		return nil, errors.New("unsupported codec type")
	}

	if err != nil {
		return nil, err
	}

	t := &TrackWriter{
		sb:       sb,
		pipeline: pipeline,
		track:    track,
	}
	go t.start()
	return t, nil
}

func (t *TrackWriter) start() {
	t.pipeline.Start()
	buf := make([]byte, 1400)

	for {
		i, _, readErr := t.track.Read(buf)

		if readErr != nil {
			panic(readErr)
		}

		t.pipeline.Push(buf[:i])

		// pkt, _, err := t.track.ReadRTP()
		// if err != nil {
		// 	break
		// }
		// t.sb.Push(pkt)

		// for _, p := range t.sb.PopPackets() {

		// 	opusPacket := codecs.OpusPacket{}
		// 	if _, err := opusPacket.Unmarshal(p.Payload); err != nil {
		// 		payload := opusPacket.Payload[0:]
		// 		t.pipeline.Push(payload)
		// 	}
		// 	// p.Unmarshal(p.Payload)
		// 	// t.pipeline.Push(p.Payload[0:])

		// }
	}
}
