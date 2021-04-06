package tello

import (
	"bytes"
	"time"

	"github.com/pion/webrtc"
	"github.com/pion/webrtc/pkg/media"
	"github.com/pion/webrtc/pkg/media/h264reader"
)

// Reads NALs of Drone saves the newest in "latestNals"
func (t *Tellodrone) checkNALs() {
	for {
		frame, err := t.h264stream.NextNAL()
		if err != nil {
			continue
		}
		if len(t.lastNALs) > 5 {
			t.lastNALs = append(t.lastNALs[1:], frame)
		} else {
			t.lastNALs = append(t.lastNALs, frame)
		}
		t.lastNALNum++
	}
}

// Send Video to webrtc track
func (t *Tellodrone) SendToTrack(track webrtc.TrackLocalStaticSample, stop *bool) {
	timeDifference := time.Millisecond * time.Duration(1000/(t.Fps+5))
	oldFrameCount := 0
	var processFrames []*h264reader.NAL
	for {
		if *stop {
			break
		}
		var frame *h264reader.NAL
		// wait till frame changed
		for oldFrameCount == t.lastNALNum && len(processFrames) == 0 {
		}

		// append new frames
		if oldFrameCount != t.lastNALNum {
			diff := t.lastNALNum - oldFrameCount
			oldFrameCount = t.lastNALNum
			data := t.lastNALs
			if diff > len(data) {
				processFrames = append(processFrames, data[len(data)-1])
			} else {
				processFrames = append(processFrames, data[len(data)-diff:]...)
			}
		}
		frame = processFrames[0]
		processFrames = processFrames[1:]

		data := frame.Data
		// DEBUG
		// fmt.Println(frame.UnitType, oldFrameCount)
		if frame.UnitType == h264reader.NalUnitTypeCodedSliceIdr || frame.UnitType == h264reader.NalUnitTypeCodedSliceNonIdr {
			if err := track.WriteSample(media.Sample{Data: data, Duration: timeDifference}); err != nil {
				panic(err)
			}
		} else if frame.UnitType == h264reader.NalUnitTypeSPS || frame.UnitType == h264reader.NalUnitTypePPS {
			if err := track.WriteSample(media.Sample{Data: data, Duration: 0}); err != nil {
				panic(err)
			}
		}
	}

}

// replacement of h264reader
func (t *Tellodrone) receiveNAL() error {
	// buf has max. 1460
	bufOut := new(bytes.Buffer)
	for {
		buf := make([]byte, 4096)
		n, _, err := t.vidConn.ReadFromUDP(buf)
		if err != nil || n <= 0 {
			return err
		}
		inp := buf[0:n]

		if inp[0] == 0x00 && inp[1] == 0x00 && inp[2] == 0x00 && inp[3] == 0x01 {
			if bufOut.Len() > 5 {
				unit := bufOut.Bytes()[4] & 31
				t.nalCh <- NAL{
					Data:           bufOut.Bytes()[4:],
					DataWithPrefix: bufOut.Bytes(),
					UnitType:       int(unit),
				}
			}
			bufOut.Reset()
		}
		bufOut.Write(inp)
	}
}
