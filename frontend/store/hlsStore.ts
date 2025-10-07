import { create } from "zustand";
import { persist, createJSONStorage } from "zustand/middleware";

interface HlsState {
  title: string;
  playbackUrl: string;
  isLive: boolean;
  buffering: boolean;
  error: string | null;
}

interface HlsStoreState {
  hls: HlsState | null;
  
  initializeHls: (title: string, playbackUrl: string) => void;
  updateHls: (patch: Partial<HlsState>) => void;
  setLiveStatus: (isLive: boolean) => void;
  setBuffering: (buffering: boolean) => void;
  setError: (error: string | null) => void;
  clearHls: () => void;
}

export const useHlsStore = create<HlsStoreState>()(
  persist(
    (set) => ({
      hls: null,
      
      initializeHls: (title, playbackUrl) =>
        set({
          hls: {
            title,
            playbackUrl,
            isLive: false,
            buffering: false,
            error: null,
          },
        }),
      
      updateHls: (patch) =>
        set((state) =>
          state.hls ? { hls: { ...state.hls, ...patch } } : state
        ),
      
      setLiveStatus: (isLive) =>
        set((state) =>
          state.hls ? { hls: { ...state.hls, isLive } } : state
        ),
      
      setBuffering: (buffering) =>
        set((state) =>
          state.hls ? { hls: { ...state.hls, buffering } } : state
        ),
      
      setError: (error) =>
        set((state) =>
          state.hls ? { hls: { ...state.hls, error } } : state
        ),
      
      clearHls: () => set({ hls: null }),
    }),
    {
      name: "hls-storage",
      storage: createJSONStorage(() => sessionStorage),
    }
  )
);