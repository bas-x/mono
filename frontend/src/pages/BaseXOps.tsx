import { useEffect, useState } from 'react';

import { Card, MapPanel, Navbar, TimelinePanel } from '@/features';
import {
  useApi,
  useSimulationStream,
  type HealthPingResult,
  type SimulationEventEnvelope,
} from '@/lib/api';

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

export function BaseXOps() {
  const { clients, config } = useApi();
  const simulationStream = useSimulationStream();

  const [latestEvent, setLatestEvent] = useState<SimulationEventEnvelope | null>(null);
  const [pingState, setPingState] = useState<PingState>({
    status: 'idle',
    message: 'Ready',
  });

  useEffect(() => {
    const unsubscribe = simulationStream.subscribe((event) => {
      setLatestEvent(event);
    });

    return () => {
      unsubscribe();
    };
  }, [simulationStream]);

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
    <div className="flex min-h-screen flex-col gap-4 p-4">
      <a
        href="#main-content"
        className="sr-only rounded-md bg-surface px-3 py-2 text-text focus:not-sr-only focus:absolute focus:left-4 focus:top-4 focus:z-50 focus:outline-none focus-visible:ring-2 focus-visible:ring-accent focus-visible:ring-offset-2 focus-visible:ring-offset-bg"
      >
        Skip to main content
      </a>
      <Navbar title="bas X" />
      <main id="main-content" className="grid flex-1 grid-cols-1 gap-4 min-[900px]:grid-cols-[2fr_1fr]">
        <MapPanel />
        <TimelinePanel />
        <Card className="min-h-auto min-[900px]:col-span-2" ariaLabel="API status section" title="API Status">
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
              <strong>Stream state:</strong> {simulationStream.state}
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
            {pingState.status === 'success' ? <p className="text-text-muted">Time: {pingState.payload.time}</p> : null}
          </div>
        </Card>
      </main>
    </div>
  );
}
