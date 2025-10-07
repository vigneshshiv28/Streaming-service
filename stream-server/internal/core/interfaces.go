package core

import (
	"github.com/pion/webrtc/v4"
	"github.com/rs/zerolog"
)

type RTCConnection interface {
	HandleSDPOffer(offer webrtc.SessionDescription, logger *zerolog.Logger) (webrtc.SessionDescription, error)
	HandleSDPAnswer(answer webrtc.SessionDescription) error
	HandleICE(candidate webrtc.ICECandidateInit, logger *zerolog.Logger) error

	Close(logger *zerolog.Logger) error
	GetPeerConnection() *webrtc.PeerConnection
}

type RTCEventHandler interface {
	OnICECandidate(candidate *webrtc.ICECandidate, logger *zerolog.Logger) error
	ForwardTracks(track *webrtc.TrackRemote, participantID string, participantName string, kind string, clientTrackID string, receiver *webrtc.RTPReceiver, logger *zerolog.Logger) error
}

type Connection interface {
	Send([]byte) error
	Close()
	Read() ([]byte, error)
}
