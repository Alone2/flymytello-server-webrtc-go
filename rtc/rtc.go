package rtc

import (
	"github.com/pion/webrtc"

	"flymytello-server-webrtc-go/tello"
)

// Start WebRTC Connection and send Video
func InitializeRTCVideo(sig Signaler, drone *tello.Tellodrone) error {
	offer, err := sig.GetSessionDescription()
	if err != nil {
		return err
	}

	// Create a MediaEngine object to configure the supported codec
	mediaEngine := webrtc.MediaEngine{}

	// define datachannel
	var dataChan *webrtc.DataChannel

	// https://raw.githubusercontent.com/chrisuehlinger/pion-h264-repro/main/pion-h264-server.go
	// Setup the codecs
	if err := mediaEngine.RegisterCodec(webrtc.RTPCodecParameters{
		RTPCodecCapability: webrtc.RTPCodecCapability{
			MimeType: "video/h264",
		},
		// 102 -> h264
		PayloadType: 102,
	}, webrtc.RTPCodecTypeVideo); err != nil {
		return err
	}

	// setup peerConnection with google stun server
	api := webrtc.NewAPI(webrtc.WithMediaEngine(&mediaEngine))
	peerConnection, err := api.NewPeerConnection(webrtc.Configuration{
		ICEServers: []webrtc.ICEServer{
			{
				URLs: []string{"stun:stun.l.google.com:19302"},
			},
		},
	})
	if err != nil {
		return err
	}

	// Setup h264 track
	track, err := webrtc.NewTrackLocalStaticSample(webrtc.RTPCodecCapability{MimeType: "video/h264"}, "video", "tello")
	if err != nil {
		return err
	}
	// and add it to Connection
	if _, err = peerConnection.AddTrack(track); err != nil {
		return err
	}

	// Send instructions to drone if delivered
	peerConnection.OnDataChannel(func(d *webrtc.DataChannel) {
		d.OnMessage(func(msg webrtc.DataChannelMessage) {
			drone.DoCmd(msg.Data)
		})
		dataChan = d
	})

	// If connection is there => Let drone send video to track
	stopSending := false
	peerConnection.OnICEConnectionStateChange(func(connState webrtc.ICEConnectionState) {
		// test if connected
		if connState == webrtc.ICEConnectionStateConnected {
			// connected
			go drone.SendToTrack(*track, &stopSending)
		} else if connState == webrtc.ICEConnectionStateClosed || connState == webrtc.ICEConnectionStateDisconnected {
			// disconnected
			stopSending = true
			dataChan.Close()
		}
	})

	// Set the remote SessionDescription
	if err = peerConnection.SetRemoteDescription(offer); err != nil {
		return err
	}

	// Create answer (to remote SessionDescription)
	answer, err := peerConnection.CreateAnswer(nil)
	if err != nil {
		return err
	}

	// Set the local SessionDescription
	if err = peerConnection.SetLocalDescription(answer); err != nil {
		return err
	}

	// Send answer to client
	sig.SendSessionDescription(answer)

	// send ICE Candidates
	peerConnection.OnICECandidate(func(i *webrtc.ICECandidate) {
		if i != nil {
			sig.SendJSONICECandidate(i.ToJSON())
		} else {
			sig.Close()
		}
	})

	// no error
	return nil
}
