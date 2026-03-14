import { IoMdJet } from 'react-icons/io';
import type { MapViewBox } from '@/features/map/types';
import type { AircraftPosition } from '@/features/simulation/hooks/useSimulation';
import { useSmoothAircrafts } from '@/features/map/hooks/useSmoothAircrafts';

type AircraftOverlayLayerProps = {
  aircraftPositions?: AircraftPosition[];
  containerSize: { width: number; height: number };
  viewBox: MapViewBox;
};

const DEFAULT_AIRCRAFT_STROKE = '#94a3b8';

const AIRCRAFT_STATE_STROKE_COLORS: Record<string, string> = {
  inbound: '#2563eb',
  landing: '#f59e0b',
  servicing: '#d946ef',
  turnaround: '#06b6d4',
  ready: '#16a34a',
  holding: '#7c3aed',
  assessment: '#f97316',
  repair: '#dc2626',
  taxi: '#eab308',
};

function normalizeAircraftState(state: string) {
  return state.trim().toLowerCase();
}

export function getAircraftStrokeColor(state: string) {
  return AIRCRAFT_STATE_STROKE_COLORS[normalizeAircraftState(state)] ?? DEFAULT_AIRCRAFT_STROKE;
}

export function AircraftOverlayLayer({
  aircraftPositions,
  containerSize,
  viewBox,
}: AircraftOverlayLayerProps) {
  const renderedAircrafts = useSmoothAircrafts(aircraftPositions, containerSize, viewBox);

  if (renderedAircrafts.length === 0) {
    return null;
  }

  return (
    <div className="pointer-events-none absolute inset-0 overflow-hidden" aria-hidden="true">
      {renderedAircrafts.map((aircraft) => (
        <div
          key={aircraft.tailNumber}
          className="absolute flex items-center justify-center text-primary"
          data-aircraft-state={normalizeAircraftState(aircraft.state)}
          style={{
            left: `calc(${aircraft.x}% - 20px)`,
            top: `calc(${aircraft.y}% - 20px)`,
            transform: `rotate(${aircraft.rotation}deg)`,
            transformOrigin: 'center',
            willChange: 'left, top, transform',
            width: 40,
            height: 40,
          }}
        >
          <IoMdJet 
            className="h-10 w-10 text-white drop-shadow-md" 
            style={{ 
              stroke: getAircraftStrokeColor(aircraft.state),
              strokeWidth: '24',
              strokeLinejoin: 'round' 
            }} 
          />
        </div>
      ))}
    </div>
  );
}
