import type {
  ConnectionState,
  SimulationEventEnvelope,
  SimulationStreamClient,
  Unsubscribe,
} from '@/lib/api/types';

const MOCK_RUN_ID = 'mock-run-001';
const STEP_INTERVAL_MS = 1_500;

const SCRIPTED_EVENTS: Array<Omit<SimulationEventEnvelope, 'runId' | 'sequence' | 'timestamp'>> = [
  {
    type: 'simulation.started',
    payload: { aircraftInQueue: 3 },
  },
  {
    type: 'simulation.progress',
    payload: { completedTasks: 2, activeResources: 4 },
  },
  {
    type: 'simulation.progress',
    payload: { completedTasks: 5, activeResources: 3 },
  },
  {
    type: 'simulation.completed',
    payload: { turnaroundMinutesSaved: 18 },
  },
];

export function createMockSimulationStreamClient(): SimulationStreamClient {
  let connectionState: ConnectionState = 'idle';
  let connectTimer: ReturnType<typeof setTimeout> | null = null;
  let intervalTimer: ReturnType<typeof setInterval> | null = null;
  let sequence = 0;
  let scriptIndex = 0;

  const eventSubscribers = new Set<(event: SimulationEventEnvelope) => void>();
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
    sequence += 1;

    const event: SimulationEventEnvelope = {
      ...scriptedEvent,
      runId: MOCK_RUN_ID,
      sequence,
      timestamp: new Date(1704067200000 + sequence * 60_000).toISOString(),
    };

    eventSubscribers.forEach((handler) => {
      handler(event);
    });

    scriptIndex = (scriptIndex + 1) % SCRIPTED_EVENTS.length;
  }

  function connect() {
    if (connectTimer || intervalTimer || connectionState === 'open') {
      return;
    }

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

  function subscribe(handler: (event: SimulationEventEnvelope) => void): Unsubscribe {
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
