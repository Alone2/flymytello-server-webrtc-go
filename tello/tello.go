package tello

import (
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"strconv"

	"github.com/pion/webrtc/pkg/media/h264reader"
)

// Standartwerte für Dji Tello
const (
	messagesIp   = "192.168.10.1"
	messagesPort = 8889
	videoIp      = "0.0.0.0"
	videoPort    = 11111
	fps          = 25
)

type NAL struct {
	Data           []byte
	DataWithPrefix []byte
	UnitType       int
}

type Tellodrone struct {
	h264stream *h264reader.H264Reader
	vidConn    *net.UDPConn
	msgConn    *net.UDPConn
	nalCh      chan NAL
	Open       bool
	Fps        int
	lastNALs   []*h264reader.NAL
	lastNALNum int
}

type inputInstructions struct {
	Command           string
	Forwardsbackwards float32
	Updown            float32
	Leftright         float32
	Yaw               float32
}

// Ganerate new TellNewTellodrone struct
func NewTellodrone() (*Tellodrone, error) {
	// Message: New address struct
	addr := net.UDPAddr{
		Port: messagesPort,
		IP:   net.ParseIP(messagesIp),
	}
	// New Connection
	connMsg, err := net.DialUDP("udp", nil, &addr)
	if err != nil {
		return &Tellodrone{}, err
	}

	// Video: New address struct
	addr = net.UDPAddr{
		Port: videoPort,
		IP:   net.ParseIP(videoIp),
	}

	// New Connection
	connVid, err := net.ListenUDP("udp", &addr)
	if err != nil {
		return &Tellodrone{}, err
	}

	// return new struct
	t := Tellodrone{
		vidConn:    connVid,
		Open:       true,
		Fps:        fps,
		nalCh:      make(chan NAL, 200),
		msgConn:    connMsg,
		lastNALNum: 0,
	}

	// Verbinde mit Drohne
	t.SendMsg("command")
	t.SendMsg("streamon")
	// Reader
	t.h264stream, err = h264reader.NewReader(&t)

	go t.checkNALs()
	return &t, nil
}

// Close connections
func (t *Tellodrone) Close() {
	t.vidConn.Close()
	t.Open = false
}

// io.Reader interface implementiert
func (t Tellodrone) Read(p []byte) (n int, err error) {
	if !t.Open {
		return 0, errors.New("Connection closed, use NewTellodrone for new connection")
	}
	n2, _, err2 := t.vidConn.ReadFromUDP(p)
	p = p[:n2]
	return n2, err2
}

// Drone executes instuction received
func (t *Tellodrone) DoCmd(msg []byte) error {
	inp := inputInstructions{}
	if err := json.Unmarshal(msg, &inp); err != nil {
		return err
	}

	switch inp.Command {

	case "fly":
		a := strconv.Itoa(int(inp.Forwardsbackwards * 100))
		b := strconv.Itoa(int(inp.Leftright * 100))
		c := strconv.Itoa(int(inp.Updown * 100))
		d := strconv.Itoa(int(inp.Yaw * 100))
		t.SendMsg("rc " + b + " " + a + " " + c + " " + d)

	default:
		go t.SendMsg(inp.Command)
	}
	return nil
}

// Cycle trough h264 NALs
func (t *Tellodrone) NextNAL() NAL {
	p := <-t.nalCh
	return p
}

// Send SDK Command to drone
func (t *Tellodrone) SendMsg(msg string) {
	fmt.Println("sending: " + msg)
	fmt.Fprintf(t.msgConn, msg)
}
