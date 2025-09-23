package streaming

import (
	"stream-server/internal/core"
	"time"

	"github.com/pion/rtcp"
	"github.com/pion/webrtc/v4"
	"github.com/rs/zerolog"
)

func (r *Room) AddTrack(track *webrtc.TrackRemote) *webrtc.TrackLocalStaticRTP {
	r.mu.Lock()
	defer func() {
		r.mu.Unlock()
	}()

	trackLocal, err := webrtc.NewTrackLocalStaticRTP(track.Codec().RTPCodecCapability, track.ID(), track.StreamID())

	if err != nil {

	}

	r.trackLocals[track.ID()] = trackLocal
	return trackLocal
}

func (r *Room) dispatchKeyFrame() {
	r.mu.Lock()
	defer r.mu.Unlock()

	for _, participant := range r.Participants {
		if participant.Role == "audience" || participant.rtcConn == nil {
			continue
		}
		peerConnection := participant.rtcConn.GetPeerConnection()
		for _, receiver := range peerConnection.GetReceivers() {
			if receiver.Track() == nil {
				continue
			}

			_ = peerConnection.WriteRTCP([]rtcp.Packet{
				&rtcp.PictureLossIndication{
					MediaSSRC: uint32(receiver.Track().SSRC()),
				},
			})
		}
	}
}

func (r *Room) SignalPeerConnections(logger *zerolog.Logger) {
	r.mu.Lock()
	defer func() {
		r.mu.Unlock()
		r.dispatchKeyFrame()
	}()

	attemptSync := func() bool {

		for _, participant := range r.Participants {

			if participant.Role == "audience" || participant.rtcConn == nil {
				continue
			}
			peerConnection := participant.rtcConn.GetPeerConnection()

			if peerConnection.ConnectionState() == webrtc.PeerConnectionStateClosed ||
				peerConnection.ConnectionState() == webrtc.PeerConnectionStateFailed {
				return true
			}

			existingSender := map[string]bool{}

			for _, sender := range peerConnection.GetSenders() {

				if sender.Track() == nil {
					continue
				}
				existingSender[sender.Track().ID()] = true

				if _, ok := r.trackLocals[sender.Track().ID()]; !ok {

					logger.Debug().Str("participant_id", participant.ID).Str("track_id", sender.Track().ID()).Msg("removing stale track")
					if err := peerConnection.RemoveTrack(sender); err != nil {
						return true
					}
				}
			}

			for _, receiver := range peerConnection.GetReceivers() {

				if receiver.Track() == nil {
					continue
				}
				existingSender[receiver.Track().ID()] = true
			}

			for trackId := range r.trackLocals {
				if _, ok := existingSender[trackId]; !ok {
					peerConnection.AddTrack(r.trackLocals[trackId])
				}
			}

			offer, err := peerConnection.CreateOffer(nil)
			if err != nil {
				logger.Error().Err(err).Str("participant_id", participant.ID).Msg("failed to create offer")
				continue
			}

			if err := peerConnection.SetLocalDescription(offer); err != nil {
				logger.Error().Err(err).Str("participant_id", participant.ID).Msg("failed to set local description")
				continue
			}

			offerMessage := core.Message{
				Type: "sdp",
				SDP:  offer,
			}

			participant.Room.SendBack(participant.ID, offerMessage, logger)

		}

		return false

	}

	for syncAttempt := 0; ; syncAttempt++ {
		if syncAttempt == 25 {

			go func() {
				time.Sleep(time.Second * 3)
				r.SignalPeerConnections(logger)
			}()

			return
		}

		if !attemptSync() {
			break
		}
	}
}
