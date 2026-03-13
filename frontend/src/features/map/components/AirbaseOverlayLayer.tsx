import { memo } from 'react';

import {
  AirbaseMarkerIcon,
  type AirbaseMarkerVariant,
} from '@/features/map/components/AirbaseMarkerIcon';
import {
  projectPointToPercent,
  type RenderContainerSize,
} from '@/features/map/lib/geometry';
import type { AirbasePoint, MapViewBox } from '@/features/map/types';

export type AirbaseCapacity = 'small' | 'medium' | 'large';

export type RenderableAirbase = {
  id: string;
  infoUrl?: string;
  centroid: AirbasePoint;
  regionId?: string;
  ariaLabel: string;
  markerSizePx: number;
};

type AirbaseOverlayLayerProps = {
  airbases: RenderableAirbase[];
  hoveredId: string | null;
  selectedId: string | null;
  containerSize: RenderContainerSize;
  viewBox: MapViewBox;
  onHoverChange: (airbase: RenderableAirbase | null) => void;
  onSelect: (airbaseId: string) => void;
};

function resolveMarkerVariant(isSelected: boolean, isHovered: boolean): AirbaseMarkerVariant {
  if (isSelected) {
    return 'selected';
  }

  if (isHovered) {
    return 'hovered';
  }

  return 'default';
}

function AirbaseOverlayLayerComponent({
  airbases,
  hoveredId,
  selectedId,
  containerSize,
  viewBox,
  onHoverChange,
  onSelect,
}: AirbaseOverlayLayerProps) {
  return (
    <div className="pointer-events-none absolute inset-0 z-10" aria-label="Airbase overlays">
      {airbases.map((airbase) => {
        const isHovered = hoveredId === airbase.id;
        const isSelected = selectedId === airbase.id;
        const markerVariant = resolveMarkerVariant(isSelected, isHovered);
        const markerPosition = projectPointToPercent(airbase.centroid, viewBox, containerSize);
        const haloSize = airbase.markerSizePx + 10;

        return (
          <button
            key={airbase.id}
            type="button"
            aria-label={airbase.ariaLabel}
            aria-pressed={isSelected}
            className="pointer-events-auto absolute flex -translate-x-1/2 -translate-y-1/2 cursor-pointer items-center justify-center rounded-full border-0 bg-transparent p-0 outline-none transition-transform hover:scale-105 focus-visible:scale-105"
            style={{
              left: `${markerPosition.x}%`,
              top: `${markerPosition.y}%`,
              width: haloSize,
              height: haloSize,
              lineHeight: 0,
            }}
            onPointerEnter={() => onHoverChange(airbase)}
            onPointerLeave={() => onHoverChange(null)}
            onFocus={() => onHoverChange(airbase)}
            onBlur={() => onHoverChange(null)}
            onClick={() => onSelect(airbase.id)}
          >
            <AirbaseMarkerIcon variant={markerVariant} sizePx={airbase.markerSizePx} />
          </button>
        );
      })}
    </div>
  );
}

export const AirbaseOverlayLayer = memo(AirbaseOverlayLayerComponent);
