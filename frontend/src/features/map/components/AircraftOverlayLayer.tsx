import { IoMdJet } from 'react-icons/io';
import type { MapViewBox } from '@/features/map/types';
import type { AircraftPosition } from '@/features/simulation/hooks/useSimulation';
import { useSmoothAircrafts } from '@/features/map/hooks/useSmoothAircrafts';

type AircraftOverlayLayerProps = {
  aircraftPositions?: AircraftPosition[];
  containerSize: { width: number; height: number };
  viewBox: MapViewBox;
};

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
              stroke: '#f59e0b',
              strokeWidth: '24',
              strokeLinejoin: 'round' 
            }} 
          />
        </div>
      ))}
    </div>
  );
}
