import { create } from "zustand";
import { persist, createJSONStorage } from "zustand/middleware";
import { Participant, ConnectionStatus, Role} from "@/types/models";

interface ParticipantStoreState {
  currentParticipant: Participant | null;
  
  setParticipant: (participant:Participant) => void;
  updateStatus: (status: ConnectionStatus) => void;
  updateRole: (role: Role) => void;
  clearParticipant: () => void;
}

export const useParticipantStore = create<ParticipantStoreState>()(
  persist(
    (set) => ({
      currentParticipant: null,

      setParticipant: (participant) =>
        set({
          currentParticipant: participant,
        }),
      
      updateStatus: (status) =>
        set((state) =>
          state.currentParticipant
            ? {
                currentParticipant: {
                  ...state.currentParticipant,
                  status,
                },
              }
            : state
        ),

      updateRole: (role) =>
        set((state) =>
          state.currentParticipant
            ? {
                currentParticipant: {
                  ...state.currentParticipant,
                  role,
                },
              }
            : state
        ),
      
      clearParticipant: () =>
        set({
          currentParticipant: null,
        }),
    }),
    {
      name: "participant-storage",
      storage: createJSONStorage(() => sessionStorage),
    }
  )
);