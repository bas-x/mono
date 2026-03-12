import { useCallback, useEffect, useState } from 'react';

import { useApi } from '@/lib/api';
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
    } catch (error) {
      setState({
        status: 'error',
        message: error instanceof Error ? error.message : 'Failed to load simulation',
      });
    }
  }, [clients.simulation]);

  const createSimulation = useCallback(async (seed: string) => {
    setState({ status: 'creating' });
    try {
      const { id } = await clients.simulation.createBaseSimulation(seed);
      await fetchSimulations();
      await loadSimulation(id);
    } catch (error: any) {
      if (error?.status === 409 || error?.response?.status === 409 || error?.message?.includes('already exists')) {
        setState({
          status: 'error',
          message: 'base simulation already exists',
        });
        await fetchSimulations();
      } else {
        setState({
          status: 'error',
          message: error instanceof Error ? error.message : 'Failed to create simulation',
        });
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
      console.error('Failed to reset simulation', error);
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
