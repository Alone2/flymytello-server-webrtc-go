package rtc

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"net/http"
	"strconv"

	"flymytello-server-webrtc-go/security"

	"github.com/gorilla/websocket"
	"github.com/pion/webrtc"
)

type TelloWebserver struct {
	receiveSigConnection <-chan *websocket.Conn
	passwordTool         *security.PasswordStruct
}

type SignalConn struct {
	conn         *websocket.Conn
	passwordTool *security.PasswordStruct
	token        string
}

// So that other signalers can be implemented
type Signaler interface {
	GetSessionDescription() (webrtc.SessionDescription, error)
	SendSessionDescription(sd webrtc.SessionDescription) error
	SendJSONICECandidate(ice webrtc.ICECandidateInit) error
	Close() error
}

// Start WebServer, return it
func NewTelloWebserver(port int, passwordTool *security.PasswordStruct, certPublicKeyPath string, certPrivateKeyPath string) (*TelloWebserver, error) {
	address := flag.String("addr2", ":"+strconv.Itoa(port), "http service address")
	upgrader := websocket.Upgrader{}

	sigChan := make(chan *websocket.Conn)

	http.HandleFunc("/signal", func(w http.ResponseWriter, r *http.Request) {
		c, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			fmt.Println("Can't upgrade /signal")
		}
		sigChan <- c
	})

	upgrader.CheckOrigin = func(r *http.Request) bool {
		return true
	}

	go http.ListenAndServeTLS(*address, certPublicKeyPath, certPrivateKeyPath, nil)

	return &TelloWebserver{
		receiveSigConnection: sigChan,
		passwordTool:         passwordTool,
	}, nil
}

func (t TelloWebserver) GetSignalerConn() *SignalConn {
	c := <-t.receiveSigConnection
	return &SignalConn{
		conn:         c,
		passwordTool: t.passwordTool,
		token:        "",
	}
}

type SessionDescriptionAuth struct {
	webrtc.SessionDescription
	Password string
}

type SessionDescriptionAuthOut struct {
	webrtc.SessionDescription
	Token string
}

// get SessionDescription from Signaler
func (s *SignalConn) GetSessionDescription() (webrtc.SessionDescription, error) {
	sessionDescriptionAuth := SessionDescriptionAuth{}
	_, message, err := s.conn.ReadMessage()
	if err != nil {
		return webrtc.SessionDescription{}, err
	}
	if err := json.Unmarshal(message, &sessionDescriptionAuth); err != nil {
		return webrtc.SessionDescription{}, err
	}
	isCorrectPassword, err := s.passwordTool.CheckHash(sessionDescriptionAuth.Password)
	if err != nil {
		return webrtc.SessionDescription{}, err
	}
	if !isCorrectPassword {
		return webrtc.SessionDescription{}, errors.New("password invalid")
	}
	// Send Session description
	out := webrtc.SessionDescription{
		Type: sessionDescriptionAuth.Type,
		SDP:  sessionDescriptionAuth.SDP,
	}
	return out, nil
}

// send SessionDescription to client
func (s *SignalConn) SendSessionDescription(sd webrtc.SessionDescription) error {
	out := SessionDescriptionAuthOut{
		SessionDescription: sd,
		Token:              s.token,
	}
	answer, err := json.Marshal(out)
	if err != nil {
		return err
	}
	// fmt.Println("sende description")
	if err := s.conn.WriteMessage(1, answer); err != nil {
		return err
	}
	return nil
}

// send ICECandidate to Signaler
func (s *SignalConn) SendJSONICECandidate(ice webrtc.ICECandidateInit) error {
	answer, err := json.Marshal(ice)
	if err != nil {
		return err
	}
	if err := s.conn.WriteMessage(1, answer); err != nil {
		return err
	}
	return nil
}

// Close Signaler
func (s *SignalConn) Close() error {
	return s.conn.Close()
}
