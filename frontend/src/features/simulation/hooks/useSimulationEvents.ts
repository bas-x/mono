import { useEffect, useState, useCallback, useRef } from 'react';
import { useSimulationStream } from '@/lib/api/useSimulationStream';
import type { SimulationEvent } from '@/lib/api/types';

export function useSimulationEvents(simulationId?: string, isPaused: boolean = false, isIdle: boolean = false, maxEvents = 200) {
  const stream = useSimulationStream(simulationId);
  const [events, setEvents] = useState<SimulationEvent[]>([]);
  const isPausedRef = useRef(isPaused);
  const isIdleRef = useRef(isIdle);

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
        if (next.length > maxEvents) {
          return next.slice(next.length - maxEvents);
        }
        return next;
      });
    });
  }, [stream, simulationId, maxEvents]);

  const clear = useCallback(() => {
    setEvents([]);
  }, []);

  return { events, clear };
}
