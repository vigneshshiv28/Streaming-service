import { create } from "zustand";
import { persist } from "zustand/middleware";
import { User } from "@/types/models";

interface UserState {
  user: User | null;
  setUser: (user: User) => void;
  removeUser: () => void;
}

export const useUserStore = create<UserState>()(
  persist(
    (set) => ({
      user: null,
      setUser: (user) => set({ user: user }),
      removeUser: () => set({ user: null }),
    }),
    {
      name: "user-storage", 
    }
  )
);

