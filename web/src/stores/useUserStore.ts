import { create } from "zustand";
import { setAuthToken } from "../api/auth";
import { defaultUserAvatars } from "../assets/brand";
import type { UserProfile } from "../api/types";

interface UserStore {
  /** Full profile of the locally "logged-in" user, or null after logout.
   *  Carries `username`/`avatar` so existing Header/PostCard consumers stay
   *  source-compatible. */
  currentUser: UserProfile | null;
  token: string | null;
  /** Mock auth flag — true while a local user is active. Not a real session. */
  isAuthed: boolean;
  setCurrentUser: (user: UserProfile, token?: string) => void;
  updateCurrentUser: (updates: Partial<UserProfile>) => void;
  clearAuthed: () => void;
}

/** Initial mock profile — mirrors DEFAULT_USER in src/api/db.ts.
 *  Duplicated here so the store doesn't import the db module (keeps the store
 *  a pure client-state layer, per web/AGENTS.md). */
const INITIAL_USER: UserProfile = {
  username: "user_developer_1",
  nickname: "Nova_Architect",
  email: "nova@research.ai",
  avatar: defaultUserAvatars[1],
  bio: "致力于研究大型语言模型的涌现行为。热衷于 AI 伦理，并优化系统提示词以获得确定性输出。",
  role: "资深研究员",
  uid: "849201",
  joinedAt: "2023-10-12T08:00:00.000Z",
  emailVerified: true,
  preferences: {
    aiReplyNotifications: true,
    liveActivity: true,
    themePreference: "system",
  },
};

const initialToken = localStorage.getItem("ai_forum_auth_token");

export const useUserStore = create<UserStore>((set) => ({
  currentUser: INITIAL_USER,
  token: initialToken,
  isAuthed: true,
  setCurrentUser: (user, token) => {
    if (token !== undefined) setAuthToken(token);
    set({ currentUser: user, token: token ?? null, isAuthed: true });
  },
  updateCurrentUser: (updates) =>
    set((state) =>
      state.currentUser
        ? { currentUser: { ...state.currentUser, ...updates } }
        : state,
    ),
  clearAuthed: () => {
    setAuthToken(null);
    set({ currentUser: null, token: null, isAuthed: false });
  },
}));
