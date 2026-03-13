import { useState, useCallback } from 'react';
import { useApi } from '@/lib/api';
import { toast } from 'sonner';

export type SimulationStatus = 'idle' | 'running' | 'paused';

export function useSimulationControls(simulationId?: string) {
  const { clients } = useApi();
  const [status, setStatus] = useState<SimulationStatus>('idle');
  const [isLoading, setIsLoading] = useState(false);

  const start = useCallback(async () => {
    if (!simulationId) return;
    setIsLoading(true);
    try {
      await clients.simulation.startSimulation(simulationId);
      setStatus('running');
      toast.success('Simulation started');
    } catch (error) {
      toast.error('Failed to start simulation');
      console.error(error);
    } finally {
      setIsLoading(false);
    }
  }, [simulationId, clients.simulation]);

  const pause = useCallback(async () => {
    if (!simulationId) return;
    setIsLoading(true);
    try {
      await clients.simulation.pauseSimulation(simulationId);
      setStatus('paused');
      toast.success('Simulation paused');
    } catch (error) {
      toast.error('Failed to pause simulation');
      console.error(error);
    } finally {
      setIsLoading(false);
    }
  }, [simulationId, clients.simulation]);

  const resume = useCallback(async () => {
    if (!simulationId) return;
    setIsLoading(true);
    try {
      await clients.simulation.resumeSimulation(simulationId);
      setStatus('running');
      toast.success('Simulation resumed');
    } catch (error) {
      toast.error('Failed to resume simulation');
      console.error(error);
    } finally {
      setIsLoading(false);
    }
  }, [simulationId, clients.simulation]);

  return { status, isLoading, start, pause, resume };
}
