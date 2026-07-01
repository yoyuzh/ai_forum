import { create } from "zustand";

interface ConnectionStore {
  sseStatus: "connected" | "connecting" | "disconnected";
  setSSEStatus: (sseStatus: "connected" | "connecting" | "disconnected") => void;
}

export const useConnectionStore = create<ConnectionStore>((set) => ({
  sseStatus: "connected",
  setSSEStatus: (sseStatus) => set({ sseStatus })
}));
