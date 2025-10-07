import { create } from "zustand";
import { persist, createJSONStorage } from "zustand/middleware";
import { Participant } from "@/types/models";

export type TrackMetaData = {
  id: string;
  kind: "audio" | "video" | "screen";
  participantId: string;
  participantName: string;
};

interface VideoCallState {
  audioTrackId: string | null;
  videoTrackId: string | null;
  screenShareId: string | null;
  muted: boolean;
  cameraOn: boolean;
  screenSharing: boolean;

  participants: Participant[];
  participantTracks: Record<string, TrackMetaData[]>; 
}


interface VideoCallStoreState {
  videoCall: VideoCallState | null;

  initializeVideoCall: () => void;
  
  setAudioTrack: (trackId: string | null) => void;
  setVideoTrack: (trackId: string | null) => void;
  setScreenShareTrack: (trackId: string | null) => void;
  
  toggleMute: () => void;
  toggleCamera: () => void;
  setMuted: (muted: boolean) => void;
  setCameraOn: (cameraOn: boolean) => void;
  setScreenSharing: (screenSharing: boolean) => void;
  
  addParticipant: (participant: Participant) => void;
  removeParticipant: (userId: string) => void;
  updateParticipant: (userId: string, patch: Partial<Participant>) => void;
  setParticipants: (participants: Participant[]) => void;
  
  setParticipantTracks: (participantId: string, tracks: TrackMetaData[]) => void;
  addParticipantTrack: (track: TrackMetaData) => void;
  removeParticipantTrack: (participantId: string, trackId: string) => void;
  clearVideoCall: () => void;
}

export const useVideoCallStore = create<VideoCallStoreState>()(
  persist(
    (set) => ({
      videoCall: null,
      
      initializeVideoCall: () =>
        set({
          videoCall: {
            audioTrackId: null,
            videoTrackId: null,
            screenShareId: null,
            muted: false,
            cameraOn: true,
            screenSharing: false,
            participants: [],
            participantTracks: {},
          },
        }),
      
      setAudioTrack: (trackId) =>
        set((state) =>
          state.videoCall
            ? { videoCall: { ...state.videoCall, audioTrackId: trackId } }
            : state
        ),
      
      setVideoTrack: (trackId) =>
        set((state) =>
          state.videoCall
            ? { videoCall: { ...state.videoCall, videoTrackId: trackId } }
            : state
        ),
      
      setScreenShareTrack: (trackId) =>
        set((state) =>
          state.videoCall
            ? { videoCall: { ...state.videoCall, screenShareId: trackId } }
            : state
        ),
      
      toggleMute: () =>
        set((state) =>
          state.videoCall
            ? { videoCall: { ...state.videoCall, muted: !state.videoCall.muted } }
            : state
        ),
      
      toggleCamera: () =>
        set((state) =>
          state.videoCall
            ? { videoCall: { ...state.videoCall, cameraOn: !state.videoCall.cameraOn } }
            : state
        ),
      
      setMuted: (muted) =>
        set((state) =>
          state.videoCall
            ? { videoCall: { ...state.videoCall, muted } }
            : state
        ),
      
      setCameraOn: (cameraOn) =>
        set((state) =>
          state.videoCall
            ? { videoCall: { ...state.videoCall, cameraOn } }
            : state
        ),
      
      setScreenSharing: (screenSharing) =>
        set((state) =>
          state.videoCall
            ? { videoCall: { ...state.videoCall, screenSharing } }
            : state
        ),
      
      addParticipant: (participant) =>
        set((state) =>
          state.videoCall
            ? {
                videoCall: {
                  ...state.videoCall,
                  participants: [...state.videoCall.participants, participant],
                },
              }
            : state
        ),
      
      removeParticipant: (userId) =>
        set((state) =>
          state.videoCall
            ? {
                videoCall: {
                  ...state.videoCall,
                  participants: state.videoCall.participants.filter(
                    (p) => p.id !== userId
                  ),
                },
              }
            : state
        ),
      
      updateParticipant: (userId, patch) =>
        set((state) =>
          state.videoCall
            ? {
                videoCall: {
                  ...state.videoCall,
                  participants: state.videoCall.participants.map((p) =>
                    p.id === userId ? { ...p, ...patch } : p
                  ),
                },
              }
            : state
        ),
      
      setParticipants: (participants) =>
        set((state) =>
          state.videoCall
            ? { videoCall: { ...state.videoCall, participants } }
            : state
        ),

        setParticipantTracks: (participantId, tracks) =>
          set((state) => ({
            videoCall: state.videoCall
              ? {
                  ...state.videoCall,
                  participantTracks: {
                    ...state.videoCall.participantTracks,
                    [participantId]: tracks,
                  },
                }
              : state.videoCall,
          })),

        addParticipantTrack: (track) =>
          set((state) => {
            if (!state.videoCall) return state;
            const existing = state.videoCall.participantTracks[track.participantId] || [];
            return {
              videoCall: {
                ...state.videoCall,
                participantTracks: {
                  ...state.videoCall.participantTracks,
                  [track.participantId]: [...existing.filter(t => t.id !== track.id), track],
                },
              },
            };
          }),

        removeParticipantTrack: (participantId, trackId) =>
          set((state) => {
            if (!state.videoCall) return state;
            return {
              videoCall: {
                ...state.videoCall,
                participantTracks: {
                  ...state.videoCall.participantTracks,
                  [participantId]: (state.videoCall.participantTracks[participantId] || []).filter(
                    (t) => t.id !== trackId
                  ),
                },
              },
            };
          }),
      
      clearVideoCall: () => set({ videoCall: null }),
    }),
    {
      name: "video-call-storage",
      storage: createJSONStorage(() => sessionStorage),
    }
  )
);