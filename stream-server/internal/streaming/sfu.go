package streaming

import (
	"stream-server/internal/core"
	"time"

	"github.com/pion/rtcp"
	"github.com/pion/rtp"
	"github.com/pion/webrtc/v4"
	"github.com/rs/zerolog"
)

func (r *Room) AddTrack(track *webrtc.TrackRemote, logger *zerolog.Logger) *webrtc.TrackLocalStaticRTP {
	logger.Debug().Str("track_id", track.ID()).Str("track_kind", track.Kind().String()).Msg("Adding track to room")

	r.mu.Lock()
	trackLocal, err := webrtc.NewTrackLocalStaticRTP(track.Codec().RTPCodecCapability, track.ID(), track.StreamID())

	logger.Debug().Str("track_id", trackLocal.ID()).Msg("Adding new track to the map")
	if err != nil {
		logger.Error().Err(err).Msg("failed to create local track")
		r.mu.Unlock()
		r.SignalPeerConnections(logger)
		return nil
	}

	r.trackLocals[track.ID()] = trackLocal
	r.mu.Unlock()

	logger.Debug().Str("track_id", track.ID()).Str("track_kind", track.Kind().String()).Msg("Added track to room")

	return trackLocal
}

func (r *Room) RemoveTrack(track *webrtc.TrackLocalStaticRTP, logger *zerolog.Logger) {
	r.mu.Lock()
	delete(r.trackLocals, track.ID())
	r.mu.Unlock()

	r.SignalPeerConnections(logger)
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

	outgoingTracks := r.GetTracksUnlocked(logger)

	attemptSync := func() bool {
		logger.Debug().Str("room_id", r.ID).Msg("Attempting to sync peer connections")
		for _, participant := range r.Participants {

			if participant.Role == "audience" || participant.rtcConn == nil {
				continue
			}
			peerConnection := participant.rtcConn.GetPeerConnection()

			logger.Debug().Str("room_id", r.ID).Str("participant_id", participant.ID).Str("pc_state", peerConnection.ConnectionState().String()).Msg("Syncing peer connection")

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

					logger.Debug().Str("room_id", r.ID).Str("participant_id", participant.ID).Str("track_id", sender.Track().ID()).Msg("removing stale track")
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

			for trackID := range r.trackLocals {
				if _, ok := existingSender[trackID]; !ok {
					if _, err := peerConnection.AddTrack(r.trackLocals[trackID]); err != nil {
						return true
					}
				}
			}

			if peerConnection.SignalingState() != webrtc.SignalingStateStable {
				logger.Warn().
					Str("participant_id", participant.ID).
					Str("state", peerConnection.SignalingState().String()).
					Msg("PeerConnection not stable, skipping offer creation")
				continue
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
				Type:           "sdp",
				SDP:            &offer,
				OutgoingTracks: outgoingTracks,
			}

			participant.Room.sendBackLocked(participant.ID, offerMessage, logger)

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

func (r *Room) GetTracks(logger *zerolog.Logger) []core.OutgoingTrackMetaData {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.GetTracksUnlocked(logger)
}

func (r *Room) GetTracksUnlocked(logger *zerolog.Logger) []core.OutgoingTrackMetaData {
	outgoingTracks := make([]core.OutgoingTrackMetaData, 0, len(r.trackMeta))

	for clientTrackID, meta := range r.trackMeta {
		if meta.TrackLocal == nil {
			logger.Warn().
				Str("client_track_id", clientTrackID).
				Msg("TrackLocal is nil for this trackMeta entry")
			continue
		}

		participant, exists := r.Participants[meta.ParticipantID]
		if !exists {
			logger.Warn().
				Str("participant_id", meta.ParticipantID).
				Msg("Participant not found for trackMeta entry")
			continue
		}

		outgoingTracks = append(outgoingTracks, core.OutgoingTrackMetaData{
			ClientTrackID:   clientTrackID,
			TrackID:         meta.TrackLocal.ID(),
			ParticipantID:   meta.ParticipantID,
			ParticipantName: participant.Name,
			Kind:            meta.Kind,
		})
	}

	return outgoingTracks
}

func (p *Participant) ForwardTracks(track *webrtc.TrackRemote, participantID string, participantName string, kind string, clientTrackID string, receiver *webrtc.RTPReceiver, logger *zerolog.Logger) error {

	trackLocal := p.Room.AddTrack(track, logger)

	p.Room.mu.Lock()
	p.Room.trackMeta[clientTrackID] = TrackMeta{
		TrackLocal:    trackLocal,
		ParticipantID: participantID,
		Kind:          kind,
	}
	logger.Debug().Msg("Added Meta Data")
	p.Room.mu.Unlock()

	p.Room.scheduleSync(logger)

	defer p.Room.RemoveTrack(trackLocal, logger)
	defer func() {
		p.Room.mu.Lock()
		delete(p.Room.trackMeta, clientTrackID)
		p.Room.mu.Unlock()
	}()

	buf := make([]byte, 1500)
	rtpPkt := &rtp.Packet{}

	for {
		i, _, err := track.Read(buf)
		if err != nil {
			return err
		}
		if err = rtpPkt.Unmarshal(buf[:i]); err != nil {

			return err
		}

		rtpPkt.Extension = false
		rtpPkt.Extensions = nil

		if err = trackLocal.WriteRTP(rtpPkt); err != nil {
			return err
		}
	}

}
