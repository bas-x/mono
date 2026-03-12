import type {
  ConnectionState,
  SimulationEvent,
  SimulationStreamClient,
  Unsubscribe,
} from '@/lib/api/types';

const MOCK_RUN_ID = 'mock-run-001';
const STEP_INTERVAL_MS = 1_500;

const SCRIPTED_EVENTS: Array<Omit<SimulationEvent, 'simulationId' | 'timestamp'>> = [
  {
    type: 'simulation_step',
    tick: 1,
  },
  {
    type: 'simulation_step',
    tick: 2,
  },
  {
    type: 'simulation_step',
    tick: 3,
  },
];

export function createMockSimulationStreamClient(): SimulationStreamClient {
  let connectionState: ConnectionState = 'idle';
  let connectTimer: ReturnType<typeof setTimeout> | null = null;
  let intervalTimer: ReturnType<typeof setInterval> | null = null;
  let scriptIndex = 0;

  const eventSubscribers = new Set<(event: SimulationEvent) => void>();
  const stateSubscribers = new Set<(state: ConnectionState) => void>();

  function emitState(nextState: ConnectionState) {
    connectionState = nextState;
    stateSubscribers.forEach((handler) => {
      handler(nextState);
    });
  }

  function clearTimers() {
    if (connectTimer) {
      clearTimeout(connectTimer);
      connectTimer = null;
    }

    if (intervalTimer) {
      clearInterval(intervalTimer);
      intervalTimer = null;
    }
  }

  function emitNextEvent() {
    const scriptedEvent = SCRIPTED_EVENTS[scriptIndex];

    const event: SimulationEvent = {
      ...scriptedEvent,
      simulationId: MOCK_RUN_ID,
      timestamp: new Date().toISOString(),
    };

    eventSubscribers.forEach((handler) => {
      handler(event);
    });

    scriptIndex = (scriptIndex + 1) % SCRIPTED_EVENTS.length;
  }

  function connect(simulationId: string) {
    if (connectTimer || intervalTimer || connectionState === 'open') {
      return;
    }

    console.log('Mock: Connecting to simulation stream', simulationId);
    emitState('connecting');

    connectTimer = setTimeout(() => {
      connectTimer = null;
      emitState('open');
      emitNextEvent();
      intervalTimer = setInterval(emitNextEvent, STEP_INTERVAL_MS);
    }, 250);
  }

  function disconnect() {
    clearTimers();
    emitState('closed');
  }

  function subscribe(handler: (event: SimulationEvent) => void): Unsubscribe {
    eventSubscribers.add(handler);
    return () => {
      eventSubscribers.delete(handler);
    };
  }

  function onConnectionStateChange(handler: (state: ConnectionState) => void): Unsubscribe {
    stateSubscribers.add(handler);
    handler(connectionState);

    return () => {
      stateSubscribers.delete(handler);
    };
  }

  return {
    connect,
    disconnect,
    subscribe,
    onConnectionStateChange,
  };
}
