import { IoMdJet } from 'react-icons/io';
import { projectPointToPercent } from '@/features/map/lib/geometry';
import type { MapViewBox } from '@/features/map/types';
import type { AircraftPosition } from '@/features/simulation/hooks/useSimulation';

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
  if (!aircraftPositions || aircraftPositions.length === 0) {
    return null;
  }

  return (
    <div className="pointer-events-none absolute inset-0 overflow-hidden" aria-hidden="true">
      {aircraftPositions.map((aircraft) => {
        const point = { x: aircraft.position.x, y: aircraft.position.y };
        const percent = projectPointToPercent(point, viewBox, containerSize);

        if (!percent) {
          return null;
        }

        return (
          <div
            key={aircraft.tailNumber}
            className="absolute flex items-center justify-center text-primary"
            style={{
              left: `calc(${percent.x}% - 12px)`,
              top: `calc(${percent.y}% - 12px)`,
              width: 24,
              height: 24,
              transition: 'left 0.5s linear, top 0.5s linear',
            }}
          >
            <IoMdJet className="h-6 w-6 text-blue-500 drop-shadow-md" />
          </div>
        );
      })}
    </div>
  );
}
