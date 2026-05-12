import { create } from 'zustand';

interface V2XMessage {
  id: string;
  carId: string;
  type: string;
  payload: string;
  timestamp: Date;
  direction: 'in' | 'out';
}

interface V2XState {
  connected: boolean;
  messages: V2XMessage[];
  ws: WebSocket | null;
  connect: () => void;
  sendMessage: (carId: string, type: string, payload: string) => void;
  addMessage: (msg: Omit<V2XMessage, 'id' | 'timestamp'>) => void;
}

export const useV2XStore = create<V2XState>((set, get) => ({
  connected: false,
  messages: [],
  ws: null,
  connect: () => {
    if (get().ws) return;
    const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
    // When using Vite dev server, requests to /api are proxied to localhost:8080
    // But for WebSockets, it's better to connect directly to the backend config
    const ws = new WebSocket(`ws://localhost:8080/api/ws`);

    ws.onopen = () => set({ connected: true });
    ws.onclose = () => set({ connected: false, ws: null });
    ws.onmessage = (event) => {
      try {
        const data = JSON.parse(event.data);
        get().addMessage({
          carId: data.carId,
          type: data.type,
          payload: data.payload,
          direction: 'in',
        });

        // Auto-reply for PoC KEM simulation
        if (data.type === 'AUTH_CHALLENGE') {
          setTimeout(() => {
            get().sendMessage(data.carId, 'KEM_REQ', 'PUBLIC_KEY_PAYLOAD...');
          }, 1000);
        }
      } catch (e) {
        console.error(e);
      }
    };
    set({ ws });
  },
  sendMessage: (carId, type, payload) => {
    const { ws, addMessage } = get();
    if (ws && ws.readyState === WebSocket.OPEN) {
      const msg = { carId, type, payload };
      ws.send(JSON.stringify(msg));
      addMessage({ ...msg, direction: 'out' });
    }
  },
  addMessage: (msg) => {
    set((state) => {
      const newMessages = [
        { ...msg, id: Math.random().toString(36).slice(2), timestamp: new Date() },
        ...state.messages,
      ].slice(0, 50); // Keep last 50
      return { messages: newMessages };
    });
  },
}));
