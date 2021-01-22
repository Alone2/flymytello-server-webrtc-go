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
        PayloadType:        102,
    }, webrtc.RTPCodecTypeVideo); err != nil {
        return err
    }

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

    track, err := webrtc.NewTrackLocalStaticSample(webrtc.RTPCodecCapability{MimeType: "video/h264"}, "video", "tello")
    if err != nil {
        return err
    }

    if _, err = peerConnection.AddTrack(track); err != nil {
        return err
    }
    peerConnection.OnDataChannel(func(d *webrtc.DataChannel) {
        d.OnMessage(func(msg webrtc.DataChannelMessage) {
            drone.DoCmd(msg.Data)
        })
        dataChan = d
    })

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

    // Create answer
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
    return nil
}
