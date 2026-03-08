import type { ApiAirbase, ApiAirbaseDetails } from '@/lib/api/types';

const SWEDEN_SVG_GEO_BOUNDS = {
  west: 11.105169,
  north: 69.061149,
  east: 24.170696,
  south: 55.338154,
} as const;

const SWEDEN_SVG_SIZE = {
  width: 345.62482,
  height: 792.52374,
} as const;

function toSvgPointFromGeo(longitude: number, latitude: number) {
  const x =
    ((longitude - SWEDEN_SVG_GEO_BOUNDS.west) /
      (SWEDEN_SVG_GEO_BOUNDS.east - SWEDEN_SVG_GEO_BOUNDS.west)) *
    SWEDEN_SVG_SIZE.width;
  const y =
    ((SWEDEN_SVG_GEO_BOUNDS.north - latitude) /
      (SWEDEN_SVG_GEO_BOUNDS.north - SWEDEN_SVG_GEO_BOUNDS.south)) *
    SWEDEN_SVG_SIZE.height;

  return { x, y };
}

function createAirbaseAreaFromGeo(longitude: number, latitude: number, width = 12, height = 12) {
  const center = toSvgPointFromGeo(longitude, latitude);
  const halfWidth = width / 2;
  const halfHeight = height / 2;

  return [
    {
      x: Math.round((center.x - halfWidth) * 100) / 100,
      y: Math.round((center.y - halfHeight) * 100) / 100,
    },
    {
      x: Math.round((center.x + halfWidth) * 100) / 100,
      y: Math.round((center.y - halfHeight + 2) * 100) / 100,
    },
    {
      x: Math.round((center.x + halfWidth - 2) * 100) / 100,
      y: Math.round((center.y + halfHeight) * 100) / 100,
    },
    {
      x: Math.round((center.x - halfWidth - 2) * 100) / 100,
      y: Math.round((center.y + halfHeight - 2) * 100) / 100,
    },
  ];
}

export const MOCK_AIRBASES: ApiAirbase[] = [
  {
    id: 'lulea',
    infoUrl: '/map/airbase/lulea',
    area: createAirbaseAreaFromGeo(22.1217, 65.5438),
  },
  {
    id: 'arlanda',
    infoUrl: '/map/airbase/arlanda',
    area: createAirbaseAreaFromGeo(17.9238, 59.6498),
  },
  {
    id: 'visby',
    infoUrl: '/map/airbase/visby',
    area: createAirbaseAreaFromGeo(18.3462, 57.6628),
  },
  {
    id: 'goteborg',
    infoUrl: '/map/airbase/goteborg',
    area: createAirbaseAreaFromGeo(12.2923, 57.6688),
  },
];

export const MOCK_AIRBASE_DETAILS: Record<string, ApiAirbaseDetails> = {
  lulea: {
    id: 'lulea',
    name: 'Lulea Airbase',
    region: 'Norrbotten',
    status: 'Operational',
    queue: 1,
  },
  arlanda: {
    id: 'arlanda',
    name: 'Arlanda Airbase',
    region: 'Stockholm',
    status: 'Busy',
    queue: 3,
  },
  visby: {
    id: 'visby',
    name: 'Visby Airbase',
    region: 'Gotland',
    status: 'Operational',
    queue: 2,
  },
  goteborg: {
    id: 'goteborg',
    name: 'Goteborg Airbase',
    region: 'Vastra Gotaland',
    status: 'Standby',
    queue: 0,
  },
};

function stripAirbasePrefix(path: string): string {
  return path.replace(/^\/?map\/airbase\//, '');
}

function normalizeLookupKey(idOrUrl: string): string {
  if (idOrUrl.startsWith('http://') || idOrUrl.startsWith('https://')) {
    try {
      const url = new URL(idOrUrl);
      return stripAirbasePrefix(url.pathname).toLowerCase();
    } catch {
      return idOrUrl.toLowerCase();
    }
  }

  return stripAirbasePrefix(idOrUrl).toLowerCase();
}

export function getMockAirbases(): ApiAirbase[] {
  return MOCK_AIRBASES.map((airbase) => ({ ...airbase, area: [...airbase.area] }));
}

export function getMockAirbaseDetails(idOrUrl: string): ApiAirbaseDetails {
  const lookupKey = normalizeLookupKey(idOrUrl);
  const details = MOCK_AIRBASE_DETAILS[lookupKey];

  if (details) {
    return { ...details };
  }

  return {
    id: lookupKey,
    name: lookupKey.toUpperCase(),
    status: 'Unknown',
    queue: 0,
  };
}
