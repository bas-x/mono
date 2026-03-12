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
            transform: `translate3d(${aircraft.x}px, ${aircraft.y}px, 0) rotate(${aircraft.rotation}deg)`,
            transformOrigin: 'center',
            willChange: 'transform',
            marginLeft: '-12px',
            marginTop: '-12px',
            width: 24,
            height: 24,
            left: 0,
            top: 0,
          }}
        >
          <IoMdJet className="h-6 w-6 text-blue-500 drop-shadow-md" />
        </div>
      ))}
    </div>
  );
}
