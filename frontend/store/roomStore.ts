import { create } from "zustand";
import { persist, createJSONStorage } from "zustand/middleware";
import { Room } from "@/types/models";

interface RoomStoreState {
  currentRoom: Room | null;
  setRoom: (room: Room) => void;
  updateRoom: (patch: Partial<Room>) => void;
  clearRoom: () => void;
}

export const useRoomStore = create<RoomStoreState>()(
  persist(
    (set) => ({
      currentRoom: null,
      
      setRoom: (room) => set({ currentRoom: room }),
      
      updateRoom: (patch) =>
        set((state) =>
          state.currentRoom
            ? { currentRoom: { ...state.currentRoom, ...patch } }
            : state
        ),
      
      clearRoom: () => set({ currentRoom: null }),
    }),
    {
      name: "room-storage",
      storage: createJSONStorage(() => sessionStorage),
    }
  )
);