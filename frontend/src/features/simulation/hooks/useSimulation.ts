import { useCallback, useState } from 'react';

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

  const createSimulation = useCallback(async (seed: string) => {
    setState({ status: 'creating' });
    try {
      const { id } = await clients.simulation.createBaseSimulation(seed);
      
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
        message: error instanceof Error ? error.message : 'Failed to create simulation',
      });
    }
  }, [clients.simulation]);

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

  return {
    state,
    createSimulation,
    refreshData,
    reset,
  };
}
