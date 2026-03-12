import { useEffect, useRef, useState } from 'react';
import type { AircraftPosition } from '@/features/simulation/hooks/useSimulation';
import { projectPointToPercent } from '@/features/map/lib/geometry';
import type { MapViewBox } from '@/features/map/types';

export type RenderedAircraft = {
  tailNumber: string;
  x: number;
  y: number;
  rotation: number;
};

function lerp(start: number, end: number, alpha: number): number {
  return start + (end - start) * alpha;
}

function lerpAngle(start: number, end: number, alpha: number): number {
  let delta = end - start;
  delta = ((delta + 180) % 360 + 360) % 360 - 180;
  return start + delta * alpha;
}

export function useSmoothAircrafts(
  aircraftPositions: AircraftPosition[] | undefined,
  containerSize: { width: number; height: number },
  viewBox: MapViewBox
): RenderedAircraft[] {
  const [renderedAircrafts, setRenderedAircrafts] = useState<RenderedAircraft[]>([]);

  const stateRef = useRef<{
    [tailNumber: string]: {
      current: { x: number; y: number; rotation: number };
      target: { x: number; y: number; rotation: number };
    };
  }>({});

  const rafRef = useRef<number>();

  useEffect(() => {
    if (!aircraftPositions || containerSize.width <= 0 || containerSize.height <= 0) {
      return;
    }

    const state = stateRef.current;
    const currentTails = new Set<string>();

    aircraftPositions.forEach((ac) => {
      currentTails.add(ac.tailNumber);

      const percent = projectPointToPercent(ac.position, viewBox, containerSize);
      if (!percent) return;

      const pxX = (percent.x / 100) * containerSize.width;
      const pxY = (percent.y / 100) * containerSize.height;

      if (!state[ac.tailNumber]) {
        state[ac.tailNumber] = {
          current: { x: pxX, y: pxY, rotation: 0 },
          target: { x: pxX, y: pxY, rotation: 0 },
        };
      } else {
        const prevTarget = state[ac.tailNumber].target;
        
        const dx = pxX - prevTarget.x;
        const dy = pxY - prevTarget.y;
        
        let newRotation = prevTarget.rotation;
        if (Math.abs(dx) > 1 || Math.abs(dy) > 1) {
          const headingDeg = Math.atan2(dy, dx) * (180 / Math.PI);
          newRotation = headingDeg - 60;
        }

        state[ac.tailNumber].target = { x: pxX, y: pxY, rotation: newRotation };
      }
    });

    Object.keys(state).forEach((tailNumber) => {
      if (!currentTails.has(tailNumber)) {
        delete state[tailNumber];
      }
    });
  }, [aircraftPositions, containerSize, viewBox]);

  useEffect(() => {
    let lastTime = performance.now();

    const loop = (time: number) => {
      const dt = time - lastTime;
      lastTime = time;
      
      const state = stateRef.current;
      const keys = Object.keys(state);

      if (keys.length > 0) {
        const alpha = Math.min(1.0, 0.15 * (dt / 16.66));
        const updatedList: RenderedAircraft[] = [];

        keys.forEach((tailNumber) => {
          const acState = state[tailNumber];
          const { current, target } = acState;

          current.x = lerp(current.x, target.x, alpha);
          current.y = lerp(current.y, target.y, alpha);
          current.rotation = lerpAngle(current.rotation, target.rotation, alpha);

          updatedList.push({
            tailNumber,
            x: current.x,
            y: current.y,
            rotation: current.rotation,
          });
        });

        setRenderedAircrafts(updatedList);
      } else {
        setRenderedAircrafts([]);
      }

      rafRef.current = requestAnimationFrame(loop);
    };

    rafRef.current = requestAnimationFrame(loop);

    return () => {
      if (rafRef.current) {
        cancelAnimationFrame(rafRef.current);
      }
    };
  }, []);

  return renderedAircrafts;
}
