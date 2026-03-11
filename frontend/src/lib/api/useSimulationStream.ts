import { useCallback, useEffect, useMemo, useState } from 'react';

import { parseApiConfigFromEnv } from '@/lib/api/config';
import { createMockSimulationStreamClient } from '@/lib/api/mock/realtime';
import { createSimulationStreamClient } from '@/lib/api/realtime/simulationStream';
import type { ConnectionState, SimulationEventEnvelope, SimulationStreamClient } from '@/lib/api/types';

export type UseSimulationStreamResult = {
  state: ConnectionState;
  connect(simulationId: string): void;
  disconnect(code?: number, reason?: string): void;
  subscribe(handler: (event: SimulationEventEnvelope) => void): () => void;
};

function createStreamClient(): SimulationStreamClient {
  const config = parseApiConfigFromEnv();

  if (config.useMock) {
    return createMockSimulationStreamClient();
  }

  return createSimulationStreamClient(config);
}

export function useSimulationStream(simulationId?: string): UseSimulationStreamResult {
  const streamClient = useMemo(() => createStreamClient(), []);
  const [state, setState] = useState<ConnectionState>('idle');

  useEffect(() => {
    const unsubscribe = streamClient.onConnectionStateChange(setState);

    if (simulationId) {
      streamClient.connect(simulationId);
    }

    return () => {
      unsubscribe();
      streamClient.disconnect(1000, 'component unmounted');
    };
  }, [streamClient, simulationId]);

  const connect = useCallback(
    (id: string) => {
      streamClient.connect(id);
    },
    [streamClient],
  );

  const disconnect = useCallback(
    (code?: number, reason?: string) => {
      streamClient.disconnect(code, reason);
    },
    [streamClient],
  );

  const subscribe = useCallback(
    (handler: (event: SimulationEventEnvelope) => void) => {
      return streamClient.subscribe(handler);
    },
    [streamClient],
  );

  return {
    state,
    connect,
    disconnect,
    subscribe,
  };
}
