import { create } from "zustand"
import { persist,createJSONStorage } from "zustand/middleware"
import { Room } from "@/types/models"

interface RoomState {
    room: Room | null;
    setRoom: (room: Room) => void;
    updateRoom: (data: Partial<Room>) =>void
    removeRoom: () => void
}


export const useRoomStore = create<RoomState>()(
  persist(
    (set) => ({
      room: null,
      setRoom: (room) => set({ room }),
      updateRoom: (data) =>
        set((state) => ({
          room: state.room ? { ...state.room, ...data } : null,
        })),
      removeRoom: () => set({ room: null }),
    }),
    {
      name: "room-storage",
      storage: createJSONStorage(() => sessionStorage),
    }
  )
)
