package rtc

import (
	"fmt"
	"stream-server/internal/core"
	"time"

	"github.com/pion/webrtc/v4"
	"github.com/rs/zerolog"
)

type PionRTCConnection struct {
	conn    *webrtc.PeerConnection
	handler core.RTCEventHandler
}

func NewPionRTCConnection(handler core.RTCEventHandler, tracksMetaData []core.IncomingTrackMetaData, logger *zerolog.Logger) (*PionRTCConnection, error) {
	config := webrtc.Configuration{
		ICEServers: []webrtc.ICEServer{
			{
				URLs: []string{"stun:stun.l.google.com:19302"},
			},
		},
	}

	pc, err := webrtc.NewPeerConnection(config)
	if err != nil {
		return nil, err
	}

	rtcConn := &PionRTCConnection{
		conn:    pc,
		handler: handler,
	}

	pc.OnICECandidate(func(candidate *webrtc.ICECandidate) {
		if candidate == nil {
			return
		}

		rtcConn.handler.OnICECandidate(candidate, logger)
	})

	pc.OnTrack(func(track *webrtc.TrackRemote, receiver *webrtc.RTPReceiver) {

		logger.Debug().Str("track_kind", track.Kind().String()).Str("track_id", track.ID()).Msg("Got remote track")

		kind := track.Kind().String()
		var participantID, participantName, clientTrackID string
		for _, trackMetaData := range tracksMetaData {
			if trackMetaData.Kind == kind {
				participantID = trackMetaData.ParticipantID
				participantName = trackMetaData.ParticipantName
				clientTrackID = trackMetaData.ClientTrackID
				logger.Debug().Str("client_track_id", clientTrackID).Str("kind", kind).Str("participant_id", participantID).Msg("forwarding track")
				break
			}
		}

		if err := handler.ForwardTracks(track, participantID, participantName, kind, clientTrackID, receiver, logger); err != nil {
			logger.Error().Err(err)
		}

	})

	return rtcConn, nil

}

func (rc *PionRTCConnection) HandleSDPOffer(sdp webrtc.SessionDescription, logger *zerolog.Logger) (webrtc.SessionDescription, error) {
	if err := rc.conn.SetRemoteDescription(sdp); err != nil {
		return webrtc.SessionDescription{}, err
	}

	answer, err := rc.conn.CreateAnswer(nil)
	if err != nil {
		return webrtc.SessionDescription{}, err
	}

	gatherComplete := webrtc.GatheringCompletePromise(rc.conn)
	if err := rc.conn.SetLocalDescription(answer); err != nil {
		return webrtc.SessionDescription{}, err
	}

	select {
	case <-gatherComplete:
	case <-time.After(5 * time.Second):
	}

	return *rc.conn.LocalDescription(), nil
}

func (rc *PionRTCConnection) HandleSDPAnswer(sdp webrtc.SessionDescription) error {
	if err := rc.conn.SetRemoteDescription(sdp); err != nil {
		return err
	}
	return nil
}
func (rc *PionRTCConnection) HandleICE(candidate webrtc.ICECandidateInit, logger *zerolog.Logger) error {
	if err := rc.conn.AddICECandidate(candidate); err != nil {
		return err
	}

	return nil

}

func (rc *PionRTCConnection) Close(logger *zerolog.Logger) error {
	if rc.conn == nil {
		return fmt.Errorf("RTC connection already nil")
	}

	logger.Info().Msg("Closing RTC connection and cleaning up resources")

	for _, sender := range rc.conn.GetSenders() {
		if sender != nil && sender.Track() != nil {
			if err := rc.conn.RemoveTrack(sender); err != nil {
				logger.Warn().Err(err).Msg("failed to remove sender track")
			}
		}
	}

	if err := rc.conn.Close(); err != nil {
		logger.Warn().Err(err).Msg("error while closing peer connection")
	}

	rc.conn = nil

	return nil
}

func (rc *PionRTCConnection) GetPeerConnection() *webrtc.PeerConnection {
	return rc.conn
}
