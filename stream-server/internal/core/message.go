package core

import "github.com/pion/webrtc/v4"

type Message struct {
	Type           string                     `json:"type"`
	From           string                     `json:"from,omitempty"`
	To             string                     `json:"to,omitempty"`
	Role           string                     `json:"role,omitempty"`
	Name           string                     `json:"name,omitempty"`
	Content        string                     `json:"content,omitempty"`
	State          []RoomState                `json:"state,omitempty"`
	SDP            *webrtc.SessionDescription `json:"sdp,omitempty"`
	ICE            *webrtc.ICECandidateInit   `json:"ice,omitempty"`
	Action         string                     `json:"action,omitempty"`
	IncomingTracks []IncomingTrackMetaData    `json:"incomingTrackMetaData,omitempty"`
	OutgoingTracks []OutgoingTrackMetaData    `json:"outgoingTrackMetaData,omitempty"`
}

type RoomState struct {
	ParticipantID   string `json:"id"`
	ParticipantName string `json:"name"`
	Role            string `json:"role"`
	Status          string `json:"status"`
}

type IncomingTrackMetaData struct {
	ClientTrackID   string `json:"id"`
	ParticipantID   string `json:"participantId"`
	ParticipantName string `json:"participantName"`
	Kind            string `json:"kind"`
}

type OutgoingTrackMetaData struct {
	ClientTrackID   string `json:"clientTrackId"`
	TrackID         string `json:"trackId"`
	ParticipantID   string `json:"participantId"`
	ParticipantName string `json:"participantName"`
	Kind            string `json:"kind"`
}
