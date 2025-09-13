import {create} from "zustand";
import { User } from "@/types/models";



interface UserState {
    user : User | null;
    setUser : (user:User) => void;
    removeUser : () => void;
}

export const useUserStore = create<UserState>((set) => ({
  user: null,
  setUser: (user) => set({ user: user }),
  removeUser: () => set({ user: null }),
}));

