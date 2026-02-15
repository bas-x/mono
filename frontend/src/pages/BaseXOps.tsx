import { useState } from 'react';

import { Card, MapPanel, Navbar, TimelinePanel } from '@/features';
import { useRpc, type HealthPingResult } from '@/lib/rpc';

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
  const { clients, config } = useRpc();
  const [pingState, setPingState] = useState<PingState>({
    status: 'idle',
    message: 'Ready',
  });

  async function handlePing() {
    setPingState({ status: 'loading', message: 'Pinging backend...' });
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

  return (
    <div className="flex min-h-screen flex-col gap-4 p-4">
      <Navbar title="bas X" />
      <main className="grid flex-1 grid-cols-1 gap-4 min-[900px]:grid-cols-[2fr_1fr]">
        <MapPanel />
        <TimelinePanel />
        <Card className="panel-rpc" ariaLabel="RPC status section" title="RPC status">
          <div className="rpc-meta">
            <p>
              <strong>Mode:</strong> {config.useMock ? 'Mock' : 'Real'}
            </p>
            <p>
              <strong>Protocol:</strong> {config.protocol}
            </p>
            <p>
              <strong>Base URL:</strong> {config.baseUrl}
            </p>
          </div>
          <div className="rpc-actions">
            <button type="button" onClick={handlePing} disabled={pingState.status === 'loading'}>
              {pingState.status === 'loading' ? 'Pinging...' : 'Ping backend'}
            </button>
            <p className={`rpc-status rpc-status-${pingState.status}`}>{pingState.message}</p>
            {pingState.status === 'success' ? <p>Time: {pingState.payload.time}</p> : null}
          </div>
        </Card>
      </main>
    </div>
  );
}
