import { useCallback, useEffect, useState } from 'react';
import { toast } from 'sonner';

import { useApi } from '@/lib/api';
import type { SimulationAirbase, SimulationAircraft } from '@/lib/api/types';

export type SimulationState =
  | { status: 'idle' }
  | { status: 'creating' }
  | { status: 'running'; simulationId: string; airbases: SimulationAirbase[]; aircrafts: SimulationAircraft[] }
  | { status: 'error'; message: string };

function extractErrorMessage(error: unknown): string {
  if (typeof error === 'object' && error !== null && 'status' in error && 'body' in error) {
    const status = (error as any).status as number;
    const bodyStr = (error as any).body as string;
    let backendMsg = bodyStr;
    try {
      const parsed = JSON.parse(bodyStr);
      if (parsed && typeof parsed.message === 'string') {
        backendMsg = parsed.message;
      }
    } catch {
      backendMsg = bodyStr;
    }
    return `Error ${status}: ${backendMsg}`;
  }
  
  if (error instanceof Error) {
    return error.message;
  }
  
  return 'An unknown error occurred';
}

function getErrorStatus(error: unknown): number | undefined {
  if (typeof error === 'object' && error !== null && 'status' in error) {
    return (error as any).status as number;
  }
  return undefined;
}

export function useSimulation() {
  const { clients } = useApi();
  const [state, setState] = useState<SimulationState>({ status: 'idle' });
  const [simulations, setSimulations] = useState<Array<{ id: string }>>([]);
  const [isLoadingSimulations, setIsLoadingSimulations] = useState(false);

  const fetchSimulations = useCallback(async () => {
    setIsLoadingSimulations(true);
    try {
      const list = await clients.simulation.getSimulations();
      setSimulations(list);
    } catch (error) {
      console.error('Failed to fetch simulations', error);
    } finally {
      setIsLoadingSimulations(false);
    }
  }, [clients.simulation]);

  useEffect(() => {
    fetchSimulations().catch(() => {});
  }, [fetchSimulations]);

  const loadSimulation = useCallback(async (id: string) => {
    setState({ status: 'creating' });
    try {
      const [airbases, aircrafts] = await Promise.all([
        clients.simulation.getAirbases(id),
        clients.simulation.getAircrafts(id),
      ]);

      setState({
        status: 'running',
        simulationId: id,
        airbases,
        aircrafts,
      });
    } catch (error) {
      const errorMessage = extractErrorMessage(error);
      toast.error(errorMessage);
      setState({
        status: 'error',
        message: errorMessage,
      });
    }
  }, [clients.simulation]);

  const createSimulation = useCallback(async (seed: string) => {
    setState({ status: 'creating' });
    try {
      const { id } = await clients.simulation.createBaseSimulation(seed);
      await fetchSimulations();
      await loadSimulation(id);
    } catch (error: unknown) {
      const errorMessage = extractErrorMessage(error);
      const statusCode = getErrorStatus(error);
      
      toast.error(errorMessage);
      setState({
        status: 'error',
        message: errorMessage,
      });
      
      if (statusCode === 409) {
        await fetchSimulations();
      }
    }
  }, [clients.simulation, fetchSimulations, loadSimulation]);

  const refreshData = useCallback(async () => {
    if (state.status !== 'running') return;

    try {
      const [airbases, aircrafts] = await Promise.all([
        clients.simulation.getAirbases(state.simulationId),
        clients.simulation.getAircrafts(state.simulationId),
      ]);

      setState((current) => {
        if (current.status !== 'running') return current;
        return {
          ...current,
          airbases,
          aircrafts,
        };
      });
    } catch (error) {
      console.error('Failed to refresh simulation data', error);
    }
  }, [clients.simulation, state]);

  const reset = useCallback(() => {
    setState({ status: 'idle' });
  }, []);

  const triggerReset = useCallback(async () => {
    if (state.status !== 'running') return;
    try {
      await clients.simulation.resetSimulation(state.simulationId);
      await refreshData();
    } catch (error) {
      const errorMessage = extractErrorMessage(error);
      toast.error(errorMessage);
    }
  }, [clients.simulation, state, refreshData]);

  return {
    state,
    simulations,
    isLoadingSimulations,
    loadSimulation,
    createSimulation,
    refreshData,
    triggerReset,
    reset,
  };
}
