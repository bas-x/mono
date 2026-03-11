import { useCallback, useEffect, useLayoutEffect, useMemo, useRef, useState } from 'react';

import { useApi } from '@/lib/api';

import {
  AirbaseOverlayLayer,
  type AirbaseCapacity,
  type RenderableAirbase,
} from '@/features/map/components/AirbaseOverlayLayer';
import { AirbaseTooltip } from '@/features/map/components/AirbaseTooltip';
import { SwedenMap3DLayer } from '@/features/map/components/SwedenMap3DLayer';
import { useAirbaseDetails } from '@/features/map/hooks/useAirbaseDetails';
import { useAirbases } from '@/features/map/hooks/useAirbases';
import {
  projectPointToPercent,
  toAriaLabel,
} from '@/features/map/lib/geometry';
import { getPlacementAnchor, getPlacementBounds } from '@/features/map/lib/placement';
import { applyCoordinateTransform, resolveCoordinateTransform } from '@/features/map/lib/transform';
import {
  DEFAULT_MAP_VIEW_BOX,
  type Airbase,
  type AirbasePlacementSource,
  type MapCoordinateTransform,
  type MapDataSource,
  type MapMode,
  type MapViewBox,
} from '@/features/map/types';

function mergeClassNames(...parts: Array<string | undefined>) {
  return parts.filter(Boolean).join(' ');
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
  viewBox?: MapViewBox;
  placementSources?: AirbasePlacementSource[];
};

const AIRBASE_CAPACITY_SEQUENCE = ['small', 'medium', 'large'] as const;

function hashAirbaseId(id: string): number {
  let hash = 0;

  for (let index = 0; index < id.length; index += 1) {
    hash = (hash * 31 + id.charCodeAt(index)) >>> 0;
  }

  return hash;
}

function resolveAirbaseCapacity(id: string): AirbaseCapacity {
  return (
    AIRBASE_CAPACITY_SEQUENCE[hashAirbaseId(id) % AIRBASE_CAPACITY_SEQUENCE.length] ?? 'medium'
  );
}

function resolveMarkerSizePx(capacity: AirbaseCapacity): number {
  switch (capacity) {
    case 'small':
      return 16;
    case 'large':
      return 32;
    default:
      return 24;
  }
}

function asRenderableAirbase(
  source: AirbasePlacementSource,
  transform: MapCoordinateTransform,
): RenderableAirbase {
  const transformedAnchor = applyCoordinateTransform(getPlacementAnchor(source), transform);
  const bounds = getPlacementBounds(source);
  const capacity = resolveAirbaseCapacity(source.id);

  return {
    id: source.id,
    infoUrl:
      'infoUrl' in source && typeof source.infoUrl === 'string' ? source.infoUrl : undefined,
    centroid: transformedAnchor,
    markerSizePx: resolveMarkerSizePx(capacity),
    ariaLabel: `${toAriaLabel(source)} ${capacity} capacity (${Math.round(bounds.maxX - bounds.minX)}x${Math.round(bounds.maxY - bounds.minY)})`,
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
  viewBox = DEFAULT_MAP_VIEW_BOX,
  placementSources: externalPlacementSources,
}: ConstellationMapProps) {
  const { clients } = useApi();
  const containerRef = useRef<HTMLDivElement | null>(null);
  const [containerSize, setContainerSize] = useState({ width: 0, height: 0 });
  const [hoveredAirbaseId, setHoveredAirbaseId] = useState<string | null>(null);
  const [uncontrolledSelectedAirbaseId, setUncontrolledSelectedAirbaseId] = useState<string | null>(
    defaultSelectedAirbaseId,
  );

  const commitContainerSize = useCallback((width: number, height: number) => {
    setContainerSize((previous) => {
      if (previous.width === width && previous.height === height) {
        return previous;
      }

      return { width, height };
    });
  }, []);

  const updateContainerSize = useCallback(() => {
    const element = containerRef.current;
    if (!element) {
      return;
    }

    commitContainerSize(Math.round(element.clientWidth), Math.round(element.clientHeight));
  }, [commitContainerSize]);

  useLayoutEffect(() => {
    updateContainerSize();
  }, [updateContainerSize]);

  useEffect(() => {
    const element = containerRef.current;
    if (!element) {
      return;
    }

    updateContainerSize();

    const resizeObserver = new ResizeObserver((entries) => {
      const entry = entries[0];
      if (!entry) {
        return;
      }

      commitContainerSize(
        Math.round(entry.contentRect.width),
        Math.round(entry.contentRect.height),
      );
    });

    resizeObserver.observe(element);

    return () => {
      resizeObserver.disconnect();
    };
  }, [commitContainerSize, updateContainerSize]);

  const isSelectionControlled = selectedAirbaseId !== undefined;
  const effectiveSelectedAirbaseId = isSelectionControlled
    ? selectedAirbaseId
    : uncontrolledSelectedAirbaseId;

  const coordinateTransform = useMemo(() => resolveCoordinateTransform(transform), [transform]);
  const airbaseState = useAirbases({ mapClient: clients.map, dataSource });
  const hasExternalPlacementSources = externalPlacementSources !== undefined;
  const effectivePlacementSources = useMemo<AirbasePlacementSource[]>(
    () =>
      externalPlacementSources ??
      airbaseState.airbases.map((airbase: Airbase) => ({
        id: airbase.id,
        area: airbase.area,
        infoUrl: airbase.infoUrl,
      })),
    [externalPlacementSources, airbaseState.airbases],
  );

  const renderableAirbases = useMemo(() => {
    return effectivePlacementSources.map((airbase) => asRenderableAirbase(airbase, coordinateTransform));
  }, [effectivePlacementSources, coordinateTransform]);

  const renderableAirbaseById = useMemo(() => {
    const byId = new Map<string, RenderableAirbase>();

    for (const airbase of renderableAirbases) {
      byId.set(airbase.id, airbase);
    }

    return byId;
  }, [renderableAirbases]);

  const hoveredAirbase = hoveredAirbaseId
    ? (renderableAirbaseById.get(hoveredAirbaseId) ?? null)
    : null;

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

    return projectPointToPercent(hoveredAirbase.centroid, viewBox, containerSize);
  }, [containerSize, hoveredAirbase, viewBox]);

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

  if (!hasExternalPlacementSources && airbaseState.status === 'loading') {
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

  if (!hasExternalPlacementSources && airbaseState.status === 'error') {
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
        'relative h-full min-h-72 overflow-hidden rounded-lg border border-border bg-bg',
        className,
      )}
      data-mode={mode}
    >
      <svg
        className="block h-full w-full"
        viewBox={`${viewBox.minX} ${viewBox.minY} ${viewBox.width} ${viewBox.height}`}
        role="img"
        aria-label="Sweden constellation map with airbase overlays"
        onClick={(event) => {
          if (event.target === event.currentTarget) {
            handleSelect(null);
          }
        }}
      >
        <title>Constellation map</title>
        <SwedenMap3DLayer viewBox={viewBox} />

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
      </svg>

      <AirbaseOverlayLayer
        airbases={renderableAirbases}
        hoveredId={hoveredAirbaseId}
        selectedId={effectiveSelectedAirbaseId ?? null}
        containerSize={containerSize}
        viewBox={viewBox}
        onHoverChange={handleHoverChange}
        onSelect={(airbaseId) => handleSelect(airbaseId)}
      />

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
