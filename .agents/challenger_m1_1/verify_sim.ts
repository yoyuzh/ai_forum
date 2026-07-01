// Mock localStorage globally before any imports
const mockLocalStorage: Record<string, string> = {};
(global as any).localStorage = {
  getItem: (key: string) => mockLocalStorage[key] || null,
  setItem: (key: string, value: string) => { mockLocalStorage[key] = value; },
  removeItem: (key: string) => { delete mockLocalStorage[key]; },
  clear: () => { for (const k in mockLocalStorage) delete mockLocalStorage[k]; },
  length: 0,
  key: (index: number) => null
};

import { db } from "../../web/src/api/db";
import { runBackgroundAISimulation } from "../../web/src/sse/simulator";
import { sseEmitter } from "../../web/src/sse/emitter";

console.log("Initial posts count:", db.getPosts().length);
