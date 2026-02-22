import { afterEach, describe, expect, it, vi } from 'vitest';

import { createWebSocketClient } from '@/lib/api/realtime/socket';
import type { ConnectionState } from '@/lib/api/types';

class MockWebSocket {
  static readonly CONNECTING = 0;
  static readonly OPEN = 1;
  static readonly CLOSING = 2;
  static readonly CLOSED = 3;

  readonly url: string;
  readyState = MockWebSocket.CONNECTING;

  onopen: (() => void) | null = null;
  onmessage: ((event: { data: string }) => void) | null = null;
  onerror: (() => void) | null = null;
  onclose: (() => void) | null = null;

  constructor(url: string) {
    this.url = url;
  }

  triggerOpen() {
    this.readyState = MockWebSocket.OPEN;
    this.onopen?.();
  }

  triggerMessage(data: string) {
    this.onmessage?.({ data });
  }

  triggerError() {
    this.onerror?.();
  }

  triggerClose() {
    this.readyState = MockWebSocket.CLOSED;
    this.onclose?.();
  }

  close() {
    this.triggerClose();
  }
}

afterEach(() => {
  vi.restoreAllMocks();
  vi.useRealTimers();
});

describe('createWebSocketClient', () => {
  it('emits connection states and parsed events', () => {
    const createdSockets: MockWebSocket[] = [];
    const stateChanges: ConnectionState[] = [];
    const events: Array<{ type: string }> = [];

    const client = createWebSocketClient(
      { wsBaseUrl: 'wss://example.com' },
      {
        path: '/ws/simulation',
        parseEvent: (raw) => JSON.parse(raw) as { type: string },
        webSocketFactory: (url) => {
          const socket = new MockWebSocket(url);
          createdSockets.push(socket);
          return socket as unknown as WebSocket;
        },
      },
    );

    client.onConnectionStateChange((state) => {
      stateChanges.push(state);
    });

    client.subscribe((event) => {
      events.push(event);
    });

    client.connect();
    createdSockets[0].triggerOpen();
    createdSockets[0].triggerMessage('{"type":"simulation.started"}');

    expect(stateChanges).toContain('connecting');
    expect(stateChanges).toContain('open');
    expect(events).toEqual([{ type: 'simulation.started' }]);
  });

  it('reconnects with backoff when socket closes unexpectedly', () => {
    vi.useFakeTimers();
    vi.spyOn(Math, 'random').mockReturnValue(0.5);

    const createdSockets: MockWebSocket[] = [];

    const client = createWebSocketClient(
      { wsBaseUrl: 'wss://example.com' },
      {
        path: '/ws/simulation',
        parseEvent: () => null,
        webSocketFactory: (url) => {
          const socket = new MockWebSocket(url);
          createdSockets.push(socket);
          return socket as unknown as WebSocket;
        },
      },
    );

    client.connect();
    createdSockets[0].triggerOpen();
    createdSockets[0].triggerClose();

    expect(createdSockets).toHaveLength(1);

    vi.advanceTimersByTime(500);

    expect(createdSockets).toHaveLength(2);
  });
});
