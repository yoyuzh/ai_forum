import { useEffect } from "react";
import { sseEmitter } from "./emitter";

export function useSSE(eventType: string, callback: (data: any) => void) {
  useEffect(() => {
    const unsubscribe = sseEmitter.subscribe(eventType, callback);
    return () => unsubscribe();
  }, [eventType, callback]);
}
