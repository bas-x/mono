import { buildApiUrl } from '@/lib/api/http/client';
import type { ApiConfig, ConnectionState, Unsubscribe } from '@/lib/api/types';

type ReconnectOptions = {
  baseDelayMs: number;
  factor: number;
  maxDelayMs: number;
  jitterRatio: number;
};

type WebSocketFactory = (url: string) => WebSocket;

type WebSocketClientOptions<TEvent> = {
  path: string;
  parseEvent: (rawData: string) => TEvent | null;
  webSocketFactory?: WebSocketFactory;
  reconnect?: Partial<ReconnectOptions>;
};

type WebSocketClient<TEvent> = {
  connect(): void;
  disconnect(code?: number, reason?: string): void;
  subscribe(handler: (event: TEvent) => void): Unsubscribe;
  onConnectionStateChange(handler: (state: ConnectionState) => void): Unsubscribe;
};

const DEFAULT_RECONNECT_OPTIONS: ReconnectOptions = {
  baseDelayMs: 500,
  factor: 2,
  maxDelayMs: 10_000,
  jitterRatio: 0.2,
};

function computeReconnectDelay(attempt: number, options: ReconnectOptions): number {
  const exponentialDelay = Math.min(
    options.baseDelayMs * options.factor ** Math.max(0, attempt - 1),
    options.maxDelayMs,
  );
  const jitterRange = exponentialDelay * options.jitterRatio;
  const jitter = (Math.random() * 2 - 1) * jitterRange;
  return Math.max(0, Math.round(exponentialDelay + jitter));
}

function normalizeWsUrl(wsBaseUrl: string, path: string): string {
  return buildApiUrl(wsBaseUrl, path);
}

export function createWebSocketClient<TEvent>(
  config: Pick<ApiConfig, 'wsBaseUrl'>,
  options: WebSocketClientOptions<TEvent>,
): WebSocketClient<TEvent> {
  const reconnectOptions: ReconnectOptions = {
    ...DEFAULT_RECONNECT_OPTIONS,
    ...options.reconnect,
  };

  const socketFactory = options.webSocketFactory ?? ((url: string) => new WebSocket(url));

  let socket: WebSocket | null = null;
  let reconnectTimer: ReturnType<typeof setTimeout> | null = null;
  let reconnectAttempt = 0;
  let manuallyClosed = false;
  let connectionState: ConnectionState = 'idle';

  const eventSubscribers = new Set<(event: TEvent) => void>();
  const connectionStateSubscribers = new Set<(state: ConnectionState) => void>();

  const wsUrl = normalizeWsUrl(config.wsBaseUrl, options.path);

  function notifyConnectionState(nextState: ConnectionState) {
    connectionState = nextState;
    connectionStateSubscribers.forEach((handler) => {
      handler(nextState);
    });
  }

  function clearReconnectTimer() {
    if (reconnectTimer) {
      clearTimeout(reconnectTimer);
      reconnectTimer = null;
    }
  }

  function cleanupSocket() {
    if (!socket) {
      return;
    }

    socket.onopen = null;
    socket.onmessage = null;
    socket.onerror = null;
    socket.onclose = null;
    socket = null;
  }

  function scheduleReconnect() {
    reconnectAttempt += 1;
    notifyConnectionState('reconnecting');

    const delay = computeReconnectDelay(reconnectAttempt, reconnectOptions);
    reconnectTimer = setTimeout(() => {
      reconnectTimer = null;
      connect();
    }, delay);
  }

  function connect() {
    if (socket && (socket.readyState === WebSocket.CONNECTING || socket.readyState === WebSocket.OPEN)) {
      return;
    }

    clearReconnectTimer();
    notifyConnectionState(reconnectAttempt === 0 ? 'connecting' : 'reconnecting');

    manuallyClosed = false;
    socket = socketFactory(wsUrl);

    socket.onopen = () => {
      reconnectAttempt = 0;
      notifyConnectionState('open');
    };

    socket.onmessage = (event) => {
      if (typeof event.data !== 'string') {
        return;
      }

      try {
        const parsed = options.parseEvent(event.data);
        if (!parsed) {
          return;
        }

        eventSubscribers.forEach((handler) => {
          handler(parsed);
        });
      } catch {
        notifyConnectionState('error');
      }
    };

    socket.onerror = () => {
      notifyConnectionState('error');
    };

    socket.onclose = () => {
      cleanupSocket();

      if (manuallyClosed) {
        notifyConnectionState('closed');
        return;
      }

      scheduleReconnect();
    };
  }

  function disconnect(code?: number, reason?: string) {
    manuallyClosed = true;
    clearReconnectTimer();

    if (socket && socket.readyState !== WebSocket.CLOSED) {
      socket.close(code, reason);
    }

    cleanupSocket();
    notifyConnectionState('closed');
  }

  function subscribe(handler: (event: TEvent) => void): Unsubscribe {
    eventSubscribers.add(handler);
    return () => {
      eventSubscribers.delete(handler);
    };
  }

  function onConnectionStateChange(handler: (state: ConnectionState) => void): Unsubscribe {
    connectionStateSubscribers.add(handler);
    handler(connectionState);

    return () => {
      connectionStateSubscribers.delete(handler);
    };
  }

  return {
    connect,
    disconnect,
    subscribe,
    onConnectionStateChange,
  };
}
