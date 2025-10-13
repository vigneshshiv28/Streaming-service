import { create } from "zustand";
import { persist, createJSONStorage } from "zustand/middleware";
import { Participant } from "@/types/models";

export type TrackMetaData = {
  clientId: string;        
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
  serverTrackMap: Record<string, { participantId: string; kind: string }>; 
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
  addParticipants: (participants: Participant[]) => void;
  removeParticipant: (userId: string) => void;
  updateParticipant: (userId: string, patch: Partial<Participant>) => void;
  setParticipants: (participants: Participant[]) => void;

  updateTracks: (tracks: TrackMetaData[]) => void;
  removeParticipantTrack: (participantId: string, trackId: string) => void;

  clearAll: () => void;
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
            serverTrackMap: {},
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

      addParticipants: (participants) =>
        set((state) =>
          state.videoCall
            ? {
                videoCall: {
                  ...state.videoCall,
                  participants: [...state.videoCall.participants, ...participants],
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

        updateTracks: (tracks) =>
          set((state) => {
            if (!state.videoCall) return state;
        
            const participantTracks: Record<string, TrackMetaData[]> = {
              ...state.videoCall.participantTracks,
            };

            const serverTrackMap: Record<string, { participantId: string; kind: string }> = {
              ...state.videoCall.serverTrackMap,
            };
        
            for (const t of tracks) {
              if (!participantTracks[t.participantId]) {
                participantTracks[t.participantId] = [];
              }
        
              const existing = participantTracks[t.participantId].find(
                (track) => track.id === t.id
              );
              if (!existing) {
                participantTracks[t.participantId].push(t);
              }
        
              serverTrackMap[t.id] = {
                participantId: t.participantId,
                kind: t.kind,
              };
            }
        
            return {
              videoCall: {
                ...state.videoCall,
                participantTracks,
                serverTrackMap,
              },
            };
          }),
        
        
        removeParticipantTrack: (participantId, serverTrackId) =>
          set((state) => {
            if (!state.videoCall) return state;
        
            const participantTracks = {
              ...state.videoCall.participantTracks,
            };
        
            if (participantTracks[participantId]) {
              participantTracks[participantId] = participantTracks[participantId].filter(
                (t) => t.id !== serverTrackId
              );
            }
        
            const serverTrackMap = { ...state.videoCall.serverTrackMap };
            delete serverTrackMap[serverTrackId];
        
            return {
              videoCall: {
                ...state.videoCall,
                participantTracks,
                serverTrackMap,
              },
            };
          }),

      clearAll: () => set({ videoCall: null }),
    }),
    {
      name: "video-call-storage",
      storage: createJSONStorage(() => sessionStorage),
    }
  )
);