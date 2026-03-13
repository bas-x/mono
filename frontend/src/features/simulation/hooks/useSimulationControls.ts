import { useState, useCallback, useEffect, useMemo } from 'react';
import { useApi } from '@/lib/api';
import { toast } from 'sonner';

export type SimulationStatus = 'idle' | 'running' | 'paused' | 'resumable' | 'finished';

type SimulationControlState = {
  isRunnerActive: boolean;
  isRunnerPaused: boolean;
  tick?: number;
  untilTick?: number;
};

export function useSimulationControls(
  simulationId?: string,
  simulation?: SimulationControlState,
  onRefresh?: () => Promise<void>,
) {
  const { clients, config } = useApi();
  const [isLoading, setIsLoading] = useState(false);

  const status = useMemo<SimulationStatus>(() => {
    if (!simulationId || !simulation) {
      return 'idle';
    }

    if (simulation.isRunnerActive) {
      return simulation.isRunnerPaused ? 'paused' : 'running';
    }

    const reachedEnd =
      simulation.untilTick != null &&
      simulation.untilTick > 0 &&
      (simulation.tick ?? 0) >= simulation.untilTick;

    if (reachedEnd) {
      return 'finished';
    }

    return (simulation.tick ?? 0) > 0 ? 'resumable' : 'idle';
  }, [simulation, simulationId]);

  const refresh = useCallback(async () => {
    if (!onRefresh) return;
    await onRefresh();
  }, [onRefresh]);

  const start = useCallback(async () => {
    if (!simulationId) return;
    setIsLoading(true);
    try {
      await clients.simulation.startSimulation(simulationId);
      toast.success('Simulation started');
      await refresh();
      return true;
    } catch (error) {
      toast.error('Failed to start simulation');
      console.error(error);
      return false;
    } finally {
      setIsLoading(false);
    }
  }, [clients.simulation, refresh, simulationId]);

  const pause = useCallback(async () => {
    if (!simulationId) return;
    setIsLoading(true);
    try {
      await clients.simulation.pauseSimulation(simulationId);
      toast.success('Simulation paused');
      await refresh();
      return true;
    } catch (error) {
      toast.error('Failed to pause simulation');
      console.error(error);
      return false;
    } finally {
      setIsLoading(false);
    }
  }, [clients.simulation, refresh, simulationId]);

  const resume = useCallback(async () => {
    if (!simulationId) return;
    setIsLoading(true);
    try {
      if (status === 'resumable') {
        await clients.simulation.startSimulation(simulationId);
      } else {
        await clients.simulation.resumeSimulation(simulationId);
      }
      toast.success('Simulation resumed');
      await refresh();
      return true;
    } catch (error) {
      toast.error('Failed to resume simulation');
      console.error(error);
      return false;
    } finally {
      setIsLoading(false);
    }
  }, [clients.simulation, refresh, simulationId, status]);

  useEffect(() => {
    if (status !== 'running' || !simulationId || config.useMock) {
      return;
    }

    const pauseOnLeave = () => {
      const url = new URL('/simulations/pause', config.apiBaseUrl).toString();
      fetch(url, {
        method: 'POST',
        keepalive: true,
      }).catch(() => {});
    };

    window.addEventListener('pagehide', pauseOnLeave);
    return () => {
      window.removeEventListener('pagehide', pauseOnLeave);
    };
  }, [config.apiBaseUrl, config.useMock, simulationId, status]);

  return { status, isLoading, start, pause, resume };
}
