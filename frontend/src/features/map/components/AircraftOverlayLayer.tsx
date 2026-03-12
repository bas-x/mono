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
            left: `calc(${aircraft.x}% - 12px)`,
            top: `calc(${aircraft.y}% - 12px)`,
            transform: `rotate(${aircraft.rotation}deg)`,
            transformOrigin: 'center',
            willChange: 'left, top, transform',
            width: 24,
            height: 24,
          }}
        >
          <IoMdJet className="h-6 w-6 text-blue-500 drop-shadow-md" />
        </div>
      ))}
    </div>
  );
}
