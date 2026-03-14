import type {
  ConnectionState,
  SimulationEvent,
  SimulationStreamClient,
  Unsubscribe,
} from '@/lib/api/types';
import { createMockEventTimestamp } from '@/lib/api/mock/scenarios';
import {
  createMockSequenceEvent,
  getStoredMockSimulationScenario,
  subscribeToMockSimulationEvents,
} from '@/lib/api/mock/store';

const STEP_INTERVAL_MS = 1_500;

export function createMockSimulationStreamClient(): SimulationStreamClient {
  let connectionState: ConnectionState = 'idle';
  let connectTimer: ReturnType<typeof setTimeout> | null = null;
  let intervalTimer: ReturnType<typeof setInterval> | null = null;
  let currentSimulationId = 'base';
  let scriptedEvents = getStoredMockSimulationScenario(currentSimulationId).events;
  let scriptIndex = 0;
  let unsubscribeBranchEvents: Unsubscribe | null = null;

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

    if (unsubscribeBranchEvents) {
      unsubscribeBranchEvents();
      unsubscribeBranchEvents = null;
    }
  }

  function emitNextEvent() {
    const scriptedEvent = scriptedEvents[scriptIndex];
    if (!scriptedEvent) {
      clearTimers();
      return;
    }

    const event: SimulationEvent = {
      type: String(scriptedEvent.type),
      ...scriptedEvent,
      simulationId: currentSimulationId,
      timestamp: createMockEventTimestamp(scriptedEvent.tick ?? scriptIndex, scriptIndex),
      sequence: scriptIndex + 1,
    };

    eventSubscribers.forEach((handler) => {
      handler(event);
    });

    scriptIndex += 1;

    if (scriptIndex >= scriptedEvents.length) {
      clearTimers();
    }
  }

  function connect(simulationId: string) {
    if (connectTimer || intervalTimer || connectionState === 'open') {
      return;
    }

    currentSimulationId = simulationId;
    scriptedEvents = getStoredMockSimulationScenario(simulationId).events;
    scriptIndex = 0;

    if (simulationId === 'base') {
      unsubscribeBranchEvents = subscribeToMockSimulationEvents(simulationId, (event) => {
        eventSubscribers.forEach((handler) => {
          handler(createMockSequenceEvent(event, scriptIndex + 1000));
        });
      });
    } else {
      unsubscribeBranchEvents = subscribeToMockSimulationEvents(simulationId, (event) => {
        eventSubscribers.forEach((handler) => {
          handler(createMockSequenceEvent(event, scriptIndex + 1000));
        });
      });
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
