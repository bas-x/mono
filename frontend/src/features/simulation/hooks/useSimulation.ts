import { useCallback, useEffect, useState } from 'react';
import { toast } from 'sonner';

import { useApi } from '@/lib/api';
import { extractErrorMessage, getErrorStatus } from '@/lib/api/errors';
import type { SimulationAirbase, SimulationAircraft } from '@/lib/api/types';

export type SimulationState =
  | { status: 'idle' }
  | { status: 'creating' }
  | { status: 'running'; simulationId: string; airbases: SimulationAirbase[]; aircrafts: SimulationAircraft[] }
  | { status: 'error'; message: string };

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
      toast.success('Simulation loaded successfully');
    } catch (error) {
      const errorMessage = extractErrorMessage(error);
      toast.error(errorMessage);
      setState({
        status: 'error',
        message: errorMessage,
      });
    }
  }, [clients.simulation]);

  const createSimulation = useCallback(async (seed: string): Promise<boolean> => {
    setState({ status: 'creating' });
    try {
      const { id } = await clients.simulation.createBaseSimulation(seed);
      toast.success('Simulation created successfully');
      await fetchSimulations();
      await loadSimulation(id);
      return true;
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
      return false;
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
      toast.success('Simulation reset successfully', {
        action: {
          label: 'Undo',
          onClick: () => console.log('Undo reset not yet implemented on backend'),
        },
      });
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
