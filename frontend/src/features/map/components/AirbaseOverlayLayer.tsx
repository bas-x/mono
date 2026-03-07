import { memo } from 'react';

import type { AirbasePoint } from '@/features/map/types';

export type RenderableAirbase = {
  id: string;
  infoUrl?: string;
  polygonPoints: string;
  centroid: AirbasePoint;
  ariaLabel: string;
};

type AirbaseOverlayLayerProps = {
  airbases: RenderableAirbase[];
  hoveredId: string | null;
  selectedId: string | null;
  onHoverChange: (airbase: RenderableAirbase | null) => void;
  onSelect: (airbaseId: string) => void;
};

function AirbaseOverlayLayerComponent({
  airbases,
  hoveredId,
  selectedId,
  onHoverChange,
  onSelect,
}: AirbaseOverlayLayerProps) {
  return (
    <g aria-label="Airbase overlays">
      {airbases.map((airbase) => {
        const isHovered = hoveredId === airbase.id;
        const isSelected = selectedId === airbase.id;

        const fillClassName = isSelected
          ? 'fill-airbase-selected-fill'
          : isHovered
            ? 'fill-airbase-hover'
            : 'fill-airbase-default-fill';
        const strokeClassName = isSelected
          ? 'stroke-airbase-selected-border'
          : isHovered
            ? 'stroke-airbase-hover'
            : 'stroke-airbase-default-stroke';
        const strokeWidthClassName = isSelected
          ? 'stroke-[2.2]'
          : isHovered
            ? 'stroke-[1.8]'
            : 'stroke-[1.25]';

        return (
          <polygon
            key={airbase.id}
            points={airbase.polygonPoints}
            role="button"
            tabIndex={0}
            aria-label={airbase.ariaLabel}
            aria-pressed={isSelected}
            className={`${fillClassName} ${strokeClassName} ${strokeWidthClassName} cursor-pointer outline-none transition-colors focus-visible:stroke-[2.2] focus-visible:stroke-airbase-selected-border`}
            onPointerEnter={() => onHoverChange(airbase)}
            onPointerLeave={() => onHoverChange(null)}
            onFocus={() => onHoverChange(airbase)}
            onBlur={() => onHoverChange(null)}
            onClick={() => onSelect(airbase.id)}
            onKeyDown={(event) => {
              if (event.key !== 'Enter' && event.key !== ' ') {
                return;
              }

              event.preventDefault();
              onSelect(airbase.id);
            }}
          />
        );
      })}
    </g>
  );
}

export const AirbaseOverlayLayer = memo(AirbaseOverlayLayerComponent);
