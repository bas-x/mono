import { useEffect, useRef, useState } from 'react';
import type { AircraftPosition } from '@/features/simulation/hooks/useSimulation';
import { projectPointToPercent } from '@/features/map/lib/geometry';
import type { MapViewBox } from '@/features/map/types';

export type RenderedAircraft = {
  tailNumber: string;
  state: string;
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
      current: { x: number; y: number; rotation: number; state: string };
      target: { x: number; y: number; rotation: number; rawX: number; rawY: number; state: string };
    };
  }>({});

  const wasEmptyRef = useRef(true);

  useEffect(() => {
    if (!aircraftPositions) {
      return;
    }

    const state = stateRef.current;
    const currentTails = new Set<string>();

    aircraftPositions.forEach((ac) => {
      currentTails.add(ac.tailNumber);

      const percent = projectPointToPercent(ac.position, viewBox, containerSize);
      if (!percent) return;

      const ptX = percent.x;
      const ptY = percent.y;

      if (!state[ac.tailNumber]) {
        state[ac.tailNumber] = {
          current: { x: ptX, y: ptY, rotation: 0, state: ac.state },
          target: { x: ptX, y: ptY, rotation: 0, rawX: ac.position.x, rawY: ac.position.y, state: ac.state },
        };
      } else {
        const prevTarget = state[ac.tailNumber].target;
        
        const dx = ac.position.x - prevTarget.rawX;
        const dy = ac.position.y - prevTarget.rawY;
        
        let newRotation = prevTarget.rotation;
        if (Math.abs(dx) > 0.001 || Math.abs(dy) > 0.001) {
          const headingDeg = Math.atan2(dy, dx) * (180 / Math.PI);
          newRotation = headingDeg + 45;
        }

        state[ac.tailNumber].target = {
          x: ptX,
          y: ptY,
          rotation: newRotation,
          rawX: ac.position.x,
          rawY: ac.position.y,
          state: ac.state,
        };
      }
    });

    Object.keys(state).forEach((tailNumber) => {
      if (!currentTails.has(tailNumber)) {
        delete state[tailNumber];
      }
    });
  }, [aircraftPositions, containerSize, viewBox]);

  useEffect(() => {
    let lastTime: number | null = null;
    let frameId: number;

    const loop = (time: number) => {
      frameId = requestAnimationFrame(loop);

      if (lastTime === null) {
        lastTime = time;
        return;
      }

      const dt = Math.max(0, time - lastTime);
      lastTime = time;
      
      const state = stateRef.current;
      const keys = Object.keys(state);

      if (keys.length > 0) {
        wasEmptyRef.current = false;
        
        const rawAlpha = 0.15 * (dt / 16.66);
        const alpha = Number.isNaN(rawAlpha) ? 1.0 : Math.min(1.0, rawAlpha);
        
        const updatedList: RenderedAircraft[] = [];

        keys.forEach((tailNumber) => {
          const acState = state[tailNumber];
          if (!acState) return;

          const { current, target } = acState;

          current.x = Number.isNaN(target.x) ? 0 : lerp(current.x, target.x, alpha);
          current.y = Number.isNaN(target.y) ? 0 : lerp(current.y, target.y, alpha);
          current.rotation = Number.isNaN(target.rotation) ? 0 : lerpAngle(current.rotation, target.rotation, alpha);

          updatedList.push({
            tailNumber,
            state: target.state,
            x: current.x,
            y: current.y,
            rotation: current.rotation,
          });
        });

        setRenderedAircrafts(updatedList);
      } else {
        if (!wasEmptyRef.current) {
          wasEmptyRef.current = true;
          setRenderedAircrafts([]);
        }
      }
    };

    frameId = requestAnimationFrame(loop);

    return () => {
      cancelAnimationFrame(frameId);
    };
  }, []);

  return renderedAircrafts;
}
