import { useCallback, useEffect, useState } from 'react';
import { toast } from 'sonner';

import { useApi } from '@/lib/api';
import { extractErrorMessage, getErrorStatus } from '@/lib/api/errors';
import { useSimulationStream } from '@/lib/api/useSimulationStream';
import type { SimulationAirbase, SimulationAircraft, SimulationAircraftNeed, SimulationEvent } from '@/lib/api/types';

export type AircraftPosition = {
  tailNumber: string;
  position: { x: number; y: number };
  state: string;
  needs: SimulationAircraftNeed[];
};

export type SimulationState =
  | { status: 'idle' }
  | { status: 'creating' }
  | { 
      status: 'running'; 
      simulationId: string; 
      airbases: SimulationAirbase[]; 
      aircrafts: SimulationAircraft[];
      tick?: number;
      time?: string;
      aircraftPositions?: AircraftPosition[];
      history: Record<number, { aircraftPositions?: AircraftPosition[], aircrafts: SimulationAircraft[] }>;
      playbackTick?: number | null;
      maxTick?: number;
    }
  | { status: 'error'; message: string };

export function useSimulation() {
  const { clients } = useApi();
  const [state, setState] = useState<SimulationState>({ status: 'idle' });
  const [simulations, setSimulations] = useState<Array<{ id: string }>>([]);
  const [isLoadingSimulations, setIsLoadingSimulations] = useState(false);
  
  const stream = useSimulationStream(state.status === 'running' ? state.simulationId : undefined);

  useEffect(() => {
    if (state.status !== 'running') {
      return;
    }

    return stream.subscribe((event: SimulationEvent) => {
      if (event.type === 'simulation_step') {
        setState((current) => {
          if (current.status !== 'running') return current;
          const currentTick = event.tick as number;
          return {
            ...current,
            tick: currentTick,
            maxTick: currentTick > (current.maxTick ?? 0) ? currentTick : current.maxTick,
            time: event.timestamp,
            history: {
              ...current.history,
              [currentTick]: current.history[currentTick] || { aircrafts: current.aircrafts, aircraftPositions: current.aircraftPositions },
            }
          };
        });
      } else if (event.type === 'aircraft_state_change') {
        setState((current) => {
          if (current.status !== 'running') return current;
          const updatedAircrafts = current.aircrafts.map((a) =>
            a.tailNumber === event.tailNumber ? { ...a, ...event.aircraft } : a
          );
          const currentTick = current.tick ?? 0;
          return { 
            ...current, 
            aircrafts: updatedAircrafts,
            history: {
              ...current.history,
              [currentTick]: { ...current.history[currentTick], aircrafts: updatedAircrafts }
            }
          };
        });
      } else if (event.type === 'landing_assignment') {
        setState((current) => {
          if (current.status !== 'running') return current;
          const updatedAircrafts = current.aircrafts.map((a) =>
            a.tailNumber === event.tailNumber ? { ...a, assignedTo: event.baseId } : a
          );
          const currentTick = current.tick ?? 0;
          return { 
            ...current, 
            aircrafts: updatedAircrafts,
            history: {
              ...current.history,
              [currentTick]: { ...current.history[currentTick], aircrafts: updatedAircrafts }
            }
          };
        });
      } else if (event.type === 'all_aircraft_positions') {
        setState((current) => {
          if (current.status !== 'running') return current;
          const currentTick = event.tick ?? current.tick ?? 0;
          return { 
            ...current, 
            aircraftPositions: event.positions,
            history: {
              ...current.history,
              [currentTick]: { ...current.history[currentTick], aircraftPositions: event.positions }
            }
          };
        });
      }
    });
  }, [stream, state.status]);

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
        history: {},
        playbackTick: null,
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
      setState({ status: 'idle' });
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
  }, [clients.simulation, state]);

  const setPlaybackTick = useCallback((tick: number | null) => {
    setState((current) => {
      if (current.status !== 'running') return current;
      return { ...current, playbackTick: tick };
    });
  }, []);

  const visibleState = state.status === 'running' 
    ? {
        ...state,
        aircrafts: state.playbackTick != null && state.history[state.playbackTick] ? state.history[state.playbackTick].aircrafts : state.aircrafts,
        aircraftPositions: state.playbackTick != null && state.history[state.playbackTick] ? state.history[state.playbackTick].aircraftPositions : state.aircraftPositions,
      } 
    : state;

  return {
    state: visibleState,
    setPlaybackTick,
    simulations,
    isLoadingSimulations,
    loadSimulation,
    createSimulation,
    refreshData,
    triggerReset,
    reset,
  };
}
