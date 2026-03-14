import { memo } from 'react';

import { projectPointToPercent, type RenderContainerSize } from '@/features/map/lib/geometry';
import type { MapViewBox } from '@/features/map/types';
import type { SimulationThreat } from '@/lib/api/types';

type ThreatOverlayLayerProps = {
  threats?: SimulationThreat[];
  containerSize: RenderContainerSize;
  viewBox: MapViewBox;
};

const THREAT_DIAMETER_PX = 112;

function ThreatOverlayLayerComponent({ threats, containerSize, viewBox }: ThreatOverlayLayerProps) {
  if (!threats || threats.length === 0) {
    return null;
  }

  return (
    <div className="pointer-events-none absolute inset-0 overflow-hidden" aria-hidden="true">
      {threats.map((threat) => {
        const markerPosition = projectPointToPercent(threat.position, viewBox, containerSize);

        return (
          <div
            key={threat.id}
            className="absolute -translate-x-1/2 -translate-y-1/2"
            data-threat-id={threat.id}
            style={{
              left: `${markerPosition.x}%`,
              top: `${markerPosition.y}%`,
              width: THREAT_DIAMETER_PX,
              height: THREAT_DIAMETER_PX,
            }}
          >
            <div
              className="absolute inset-0 rounded-full"
              style={{
                background:
                  'radial-gradient(circle at 70% 50%, rgba(255, 55, 55, 0.26) 0%, rgba(255, 32, 32, 0.14) 48%, rgba(185, 10, 10, 0) 100%)',
                filter: 'blur(10px)',
              }}
            />
            <div
              className="absolute inset-[12%] rounded-full border"
              style={{
                background:
                  'radial-gradient(circle at 68% 50%, rgba(255, 72, 72, 0.36) 0%, rgba(255, 34, 34, 0.24) 38%, rgba(214, 18, 18, 0.12) 68%, rgba(160, 0, 0, 0) 100%)',
                borderColor: 'rgba(255, 96, 96, 0.4)',
                boxShadow: '0 0 28px rgba(220, 38, 38, 0.18)',
              }}
            />
          </div>
        );
      })}
    </div>
  );
}

export const ThreatOverlayLayer = memo(ThreatOverlayLayerComponent);
