type SSECallback = (data: any) => void;

class ClientEventEmitter {
  private listeners: Map<string, Set<SSECallback>> = new Map();

  public subscribe(eventType: string, callback: SSECallback): () => void {
    if (!this.listeners.has(eventType)) {
      this.listeners.set(eventType, new Set());
    }
    this.listeners.get(eventType)!.add(callback);

    return () => {
      const set = this.listeners.get(eventType);
      if (set) {
        set.delete(callback);
        if (set.size === 0) this.listeners.delete(eventType);
      }
    };
  }

  public emit(eventType: string, data: any) {
    const set = this.listeners.get(eventType);
    if (set) {
      set.forEach(callback => callback(data));
    }
    // Also emit to wildcards or generic log channel
    const generic = this.listeners.get("*");
    if (generic) {
      generic.forEach(callback => callback({ type: eventType, data }));
    }
  }
}

export const sseEmitter = new ClientEventEmitter();
