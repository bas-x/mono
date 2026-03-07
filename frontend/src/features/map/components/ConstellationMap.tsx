import { useCallback, useEffect, useMemo, useRef, useState } from 'react';

import swedenMapRaw from '@/assets/sweden.svg?raw';
import { useApi } from '@/lib/api';

import { AirbaseOverlayLayer, type RenderableAirbase } from '@/features/map/components/AirbaseOverlayLayer';
import { AirbaseTooltip } from '@/features/map/components/AirbaseTooltip';
import { useAirbaseDetails } from '@/features/map/hooks/useAirbaseDetails';
import { useAirbases } from '@/features/map/hooks/useAirbases';
import {
  calculatePolygonBounds,
  calculatePolygonCentroid,
  hasValidPolygon,
  pointToViewBoxPercent,
  polygonToPointsAttribute,
  toAriaLabel,
} from '@/features/map/lib/geometry';
import { applyTransformToPolygon, resolveCoordinateTransform } from '@/features/map/lib/transform';
import {
  DEFAULT_MAP_VIEW_BOX,
  type Airbase,
  type MapCoordinateTransform,
  type MapDataSource,
  type MapMode,
} from '@/features/map/types';

function mergeClassNames(...parts: Array<string | undefined>) {
  return parts.filter(Boolean).join(' ');
}

type ElementSize = {
  width: number;
  height: number;
};

function extractSvgInnerMarkup(svgRaw: string): string {
  const withoutMetadata = svgRaw
    .replace(/<\?xml[\s\S]*?\?>/gi, '')
    .replace(/<!--[\s\S]*?-->/gi, '');
  const match = withoutMetadata.match(/<svg[^>]*>([\s\S]*?)<\/svg>/i);
  return match ? match[1] : withoutMetadata;
}

const SWEDEN_SVG_INNER_MARKUP = extractSvgInnerMarkup(swedenMapRaw);

function toTooltipPercentPosition(
  point: { x: number; y: number },
  containerSize: ElementSize,
): { x: number; y: number } | null {
  if (containerSize.width <= 0 || containerSize.height <= 0) {
    return null;
  }

  const viewBoxAspectRatio = DEFAULT_MAP_VIEW_BOX.width / DEFAULT_MAP_VIEW_BOX.height;
  const containerAspectRatio = containerSize.width / containerSize.height;

  let renderedWidth = containerSize.width;
  let renderedHeight = containerSize.height;
  let offsetX = 0;
  let offsetY = 0;

  if (containerAspectRatio > viewBoxAspectRatio) {
    renderedWidth = containerSize.height * viewBoxAspectRatio;
    offsetX = (containerSize.width - renderedWidth) / 2;
  } else if (containerAspectRatio < viewBoxAspectRatio) {
    renderedHeight = containerSize.width / viewBoxAspectRatio;
    offsetY = (containerSize.height - renderedHeight) / 2;
  }

  const xRatio = (point.x - DEFAULT_MAP_VIEW_BOX.minX) / DEFAULT_MAP_VIEW_BOX.width;
  const yRatio = (point.y - DEFAULT_MAP_VIEW_BOX.minY) / DEFAULT_MAP_VIEW_BOX.height;

  const xPixel = offsetX + xRatio * renderedWidth;
  const yPixel = offsetY + yRatio * renderedHeight;

  return {
    x: (xPixel / containerSize.width) * 100,
    y: (yPixel / containerSize.height) * 100,
  };
}

type ConstellationMapProps = {
  mode?: MapMode;
  dataSource?: MapDataSource;
  transform?: Partial<MapCoordinateTransform>;
  selectedAirbaseId?: string | null;
  defaultSelectedAirbaseId?: string | null;
  onSelectAirbase?: (airbaseId: string | null) => void;
  onHoverAirbase?: (airbaseId: string | null) => void;
  fetchDetailsOnHover?: boolean;
  hoverDebounceMs?: number;
  detailsCacheTtlMs?: number;
  className?: string;
  showDebugOverlay?: boolean;
};

function asRenderableAirbase(airbase: Airbase, transform: MapCoordinateTransform): RenderableAirbase | null {
  const transformedPoints = applyTransformToPolygon(airbase.area, transform);
  if (!hasValidPolygon(transformedPoints)) {
    return null;
  }

  const bounds = calculatePolygonBounds(transformedPoints);
  const centroid = calculatePolygonCentroid(transformedPoints);

  return {
    id: airbase.id,
    infoUrl: airbase.infoUrl,
    polygonPoints: polygonToPointsAttribute(transformedPoints),
    centroid,
    ariaLabel: `${toAriaLabel(airbase)} (${Math.round(bounds.maxX - bounds.minX)}x${Math.round(bounds.maxY - bounds.minY)})`,
  };
}

export function ConstellationMap({
  mode = 'static',
  dataSource = 'mock',
  transform,
  selectedAirbaseId,
  defaultSelectedAirbaseId = null,
  onSelectAirbase,
  onHoverAirbase,
  fetchDetailsOnHover = true,
  hoverDebounceMs = 180,
  detailsCacheTtlMs = 60_000,
  className,
  showDebugOverlay = false,
}: ConstellationMapProps) {
  const { clients } = useApi();
  const containerRef = useRef<HTMLDivElement | null>(null);
  const [containerSize, setContainerSize] = useState<ElementSize>({ width: 0, height: 0 });
  const [hoveredAirbaseId, setHoveredAirbaseId] = useState<string | null>(null);
  const [uncontrolledSelectedAirbaseId, setUncontrolledSelectedAirbaseId] = useState<string | null>(
    defaultSelectedAirbaseId,
  );

  useEffect(() => {
    const element = containerRef.current;
    if (!element) {
      return;
    }

    const resizeObserver = new ResizeObserver((entries) => {
      const entry = entries[0];
      if (!entry) {
        return;
      }

      const width = Math.round(entry.contentRect.width);
      const height = Math.round(entry.contentRect.height);
      setContainerSize((previous) => {
        if (previous.width === width && previous.height === height) {
          return previous;
        }
        return { width, height };
      });
    });

    resizeObserver.observe(element);

    return () => {
      resizeObserver.disconnect();
    };
  }, []);

  const isSelectionControlled = selectedAirbaseId !== undefined;
  const effectiveSelectedAirbaseId = isSelectionControlled
    ? selectedAirbaseId
    : uncontrolledSelectedAirbaseId;

  const coordinateTransform = useMemo(() => resolveCoordinateTransform(transform), [transform]);
  const airbaseState = useAirbases({ mapClient: clients.map, dataSource });

  const renderableAirbases = useMemo(() => {
    return airbaseState.airbases
      .map((airbase) => asRenderableAirbase(airbase, coordinateTransform))
      .filter((airbase): airbase is RenderableAirbase => airbase !== null);
  }, [airbaseState.airbases, coordinateTransform]);

  const renderableAirbaseById = useMemo(() => {
    const byId = new Map<string, RenderableAirbase>();

    for (const airbase of renderableAirbases) {
      byId.set(airbase.id, airbase);
    }

    return byId;
  }, [renderableAirbases]);

  const hoveredAirbase = hoveredAirbaseId ? renderableAirbaseById.get(hoveredAirbaseId) ?? null : null;

  const detailsState = useAirbaseDetails({
    mapClient: clients.map,
    hoveredAirbase,
    dataSource,
    enabled: fetchDetailsOnHover,
    debounceMs: hoverDebounceMs,
    cacheTtlMs: detailsCacheTtlMs,
  });

  const hoveredPointPercent = useMemo(() => {
    if (!hoveredAirbase) {
      return null;
    }

    return (
      toTooltipPercentPosition(hoveredAirbase.centroid, containerSize) ??
      pointToViewBoxPercent(hoveredAirbase.centroid, DEFAULT_MAP_VIEW_BOX)
    );
  }, [containerSize, hoveredAirbase]);

  const handleHoverChange = useCallback(
    (airbase: RenderableAirbase | null) => {
      const nextHoveredId = airbase?.id ?? null;
      setHoveredAirbaseId(nextHoveredId);
      onHoverAirbase?.(nextHoveredId);
    },
    [onHoverAirbase],
  );

  const handleSelect = useCallback(
    (airbaseId: string | null) => {
      if (!isSelectionControlled) {
        setUncontrolledSelectedAirbaseId(airbaseId);
      }

      onSelectAirbase?.(airbaseId);
    },
    [isSelectionControlled, onSelectAirbase],
  );

  if (airbaseState.status === 'loading') {
    return (
      <div
        className={mergeClassNames(
          'flex h-full min-h-[18rem] items-center justify-center rounded-lg border border-dashed border-border bg-bg text-sm text-text-muted',
          className,
        )}
      >
        Loading constellation map…
      </div>
    );
  }

  if (airbaseState.status === 'error') {
    return (
      <div
        className={mergeClassNames(
          'flex h-full min-h-[18rem] flex-col items-center justify-center rounded-lg border border-red-300 bg-red-50/30 p-4 text-center text-sm text-red-700 dark:border-red-900 dark:bg-red-950/20 dark:text-red-400',
          className,
        )}
      >
        <p className="m-0 font-semibold">Unable to load map overlays.</p>
        <p className="m-0 mt-1">{airbaseState.message}</p>
      </div>
    );
  }

  if (renderableAirbases.length === 0) {
    return (
      <div
        className={mergeClassNames(
          'flex h-full min-h-[18rem] items-center justify-center rounded-lg border border-border bg-bg text-sm text-text-muted',
          className,
        )}
      >
        No airbases available for this scenario.
      </div>
    );
  }

  return (
    <div
      ref={containerRef}
      className={mergeClassNames(
        'relative h-full min-h-[18rem] overflow-hidden rounded-lg border border-border bg-bg',
        className,
      )}
      data-mode={mode}
    >
      <svg
        className="block h-full w-full"
        viewBox={`${DEFAULT_MAP_VIEW_BOX.minX} ${DEFAULT_MAP_VIEW_BOX.minY} ${DEFAULT_MAP_VIEW_BOX.width} ${DEFAULT_MAP_VIEW_BOX.height}`}
        role="img"
        aria-label="Sweden constellation map with airbase overlays"
        onClick={(event) => {
          if (event.target === event.currentTarget) {
            handleSelect(null);
          }
        }}
      >
        <title>Constellation map</title>
        <style>{`
          .sweden-map-layer path {
            fill: var(--color-map-surface);
            stroke: var(--color-map-boundary);
            stroke-width: 0.85;
            vector-effect: non-scaling-stroke;
            paint-order: stroke fill;
          }
        `}</style>
        <g
          className="sweden-map-layer pointer-events-none"
          aria-hidden="true"
          dangerouslySetInnerHTML={{ __html: SWEDEN_SVG_INNER_MARKUP }}
        />

        {showDebugOverlay ? (
          <g aria-hidden="true">
            <rect
              x={DEFAULT_MAP_VIEW_BOX.minX}
              y={DEFAULT_MAP_VIEW_BOX.minY}
              width={DEFAULT_MAP_VIEW_BOX.width}
              height={DEFAULT_MAP_VIEW_BOX.height}
              className="fill-transparent stroke-accent/40 stroke-[0.6]"
            />
            <line
              x1={DEFAULT_MAP_VIEW_BOX.minX}
              y1={DEFAULT_MAP_VIEW_BOX.minY + DEFAULT_MAP_VIEW_BOX.height / 2}
              x2={DEFAULT_MAP_VIEW_BOX.minX + DEFAULT_MAP_VIEW_BOX.width}
              y2={DEFAULT_MAP_VIEW_BOX.minY + DEFAULT_MAP_VIEW_BOX.height / 2}
              className="stroke-accent/30 stroke-[0.6]"
            />
            <line
              x1={DEFAULT_MAP_VIEW_BOX.minX + DEFAULT_MAP_VIEW_BOX.width / 2}
              y1={DEFAULT_MAP_VIEW_BOX.minY}
              x2={DEFAULT_MAP_VIEW_BOX.minX + DEFAULT_MAP_VIEW_BOX.width / 2}
              y2={DEFAULT_MAP_VIEW_BOX.minY + DEFAULT_MAP_VIEW_BOX.height}
              className="stroke-accent/30 stroke-[0.6]"
            />
          </g>
        ) : null}

        <AirbaseOverlayLayer
          airbases={renderableAirbases}
          hoveredId={hoveredAirbaseId}
          selectedId={effectiveSelectedAirbaseId ?? null}
          onHoverChange={handleHoverChange}
          onSelect={(airbaseId) => handleSelect(airbaseId)}
        />
      </svg>

      {hoveredAirbase && hoveredPointPercent ? (
        <AirbaseTooltip
          airbaseId={hoveredAirbase.id}
          leftPercent={hoveredPointPercent.x}
          topPercent={hoveredPointPercent.y}
          detailsState={detailsState}
        />
      ) : null}
    </div>
  );
}
