import { useEffect, useState, useCallback, useRef } from 'react';
import { useSimulationStream } from '@/lib/api/useSimulationStream';
import type { SimulationEvent } from '@/lib/api/types';

export function useSimulationEvents(
  simulationId?: string,
  isPaused: boolean = false,
  isIdle: boolean = false,
) {
  const stream = useSimulationStream(simulationId);
  const [events, setEvents] = useState<SimulationEvent[]>([]);
  const isPausedRef = useRef(isPaused);
  const isIdleRef = useRef(isIdle);
  const eventsBySimulationRef = useRef<Map<string, SimulationEvent[]>>(new Map());

  useEffect(() => {
    isPausedRef.current = isPaused;
    isIdleRef.current = isIdle;
  }, [isPaused, isIdle]);

  useEffect(() => {
    if (!simulationId) return;

    return stream.subscribe((event) => {
      if (isPausedRef.current || isIdleRef.current) return;

      setEvents((prev) => {
        const next = [...prev, event];
        if (simulationId) {
          eventsBySimulationRef.current.set(simulationId, next);
        }
        return next;
      });
    });
  }, [stream, simulationId]);

  useEffect(() => {
    if (!simulationId) {
      setEvents([]);
      return;
    }

    setEvents(eventsBySimulationRef.current.get(simulationId) ?? []);
  }, [simulationId]);

  const clear = useCallback(() => {
    if (simulationId) {
      eventsBySimulationRef.current.delete(simulationId);
    }
    setEvents([]);
  }, [simulationId]);

  return { events, clear };
}
