import { useEffect, useState } from 'react';

import {
  useApi,
  useSimulationStream,
  type HealthPingResult,
  type SimulationEventEnvelope,
} from '@/lib/api';

import { Card } from '@/features/ui';

type PingState =
  | { status: 'idle'; message: string }
  | { status: 'loading'; message: string }
  | { status: 'success'; message: string; payload: HealthPingResult }
  | { status: 'error'; message: string };

function friendlyError(error: unknown): string {
  if (error instanceof Error) {
    return error.message;
  }

  return 'Unable to reach backend. Verify backend is running and API base URL is correct.';
}

export function ApiStatusPanel() {
  const { clients, config } = useApi();
  const { state: streamState, subscribe } = useSimulationStream();

  const [latestEvent, setLatestEvent] = useState<SimulationEventEnvelope | null>(null);
  const [pingState, setPingState] = useState<PingState>({
    status: 'idle',
    message: 'Ready',
  });

  useEffect(() => {
    const unsubscribe = subscribe((event) => {
      setLatestEvent(event);
    });

    return () => {
      unsubscribe();
    };
  }, [subscribe]);

  async function handlePing() {
    setPingState({ status: 'loading', message: 'Pinging backend…' });

    try {
      const response = await clients.health.ping();
      setPingState({
        status: 'success',
        message: response.message,
        payload: response,
      });
    } catch (error) {
      setPingState({ status: 'error', message: friendlyError(error) });
    }
  }

  const pingStatusClassName =
    pingState.status === 'success'
      ? 'text-green-700 dark:text-green-400'
      : pingState.status === 'error'
        ? 'text-red-700 dark:text-red-400'
        : 'text-text-muted';

  return (
    <Card className="min-[900px]:col-span-2" ariaLabel="API status section" title="API Status">
      <div className="space-y-1 text-sm text-text-muted">
        <p>
          <strong>Mode:</strong> {config.useMock ? 'Mock' : 'Real'}
        </p>
        <p>
          <strong>API Base URL:</strong> {config.apiBaseUrl}
        </p>
        <p>
          <strong>WS Base URL:</strong> {config.wsBaseUrl}
        </p>
        <p>
          <strong>Stream state:</strong> {streamState}
        </p>
        {latestEvent ? (
          <p>
            <strong>Latest event:</strong> {latestEvent.type} #{latestEvent.sequence}
          </p>
        ) : (
          <p>
            <strong>Latest event:</strong> none yet
          </p>
        )}
      </div>

      <div className="mt-3 space-y-2 text-sm">
        <button
          type="button"
          onClick={handlePing}
          disabled={pingState.status === 'loading'}
          className="cursor-pointer rounded-md border border-border bg-primary px-3 py-1.5 font-medium text-header-text transition-colors hover:bg-primary-strong focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-accent focus-visible:ring-offset-2 focus-visible:ring-offset-surface disabled:cursor-not-allowed disabled:opacity-60"
        >
          {pingState.status === 'loading' ? 'Pinging…' : 'Ping backend'}
        </button>
        <p aria-live="polite" className={`font-semibold ${pingStatusClassName}`}>
          {pingState.message}
        </p>
        {pingState.status === 'success' ? (
          <p className="text-text-muted">Time: {pingState.payload.time}</p>
        ) : null}
      </div>
    </Card>
  );
}
