import type { ApiAirbase, ApiAirbaseDetails } from '@/lib/api/types';

function createAirbaseAreaFromSvgCenter(centerX: number, centerY: number, width = 12, height = 12) {
  const halfWidth = width / 2;
  const halfHeight = height / 2;

  return [
    {
      x: Math.round((centerX - halfWidth) * 100) / 100,
      y: Math.round((centerY - halfHeight) * 100) / 100,
    },
    {
      x: Math.round((centerX + halfWidth) * 100) / 100,
      y: Math.round((centerY - halfHeight + 2) * 100) / 100,
    },
    {
      x: Math.round((centerX + halfWidth - 2) * 100) / 100,
      y: Math.round((centerY + halfHeight) * 100) / 100,
    },
    {
      x: Math.round((centerX - halfWidth - 2) * 100) / 100,
      y: Math.round((centerY + halfHeight - 2) * 100) / 100,
    },
  ];
}

export const MOCK_AIRBASES: ApiAirbase[] = [
  {
    id: 'lulea',
    infoUrl: '/map/airbase/lulea',
    area: createAirbaseAreaFromSvgCenter(287, 178),
  },
  {
    id: 'arlanda',
    infoUrl: '/map/airbase/arlanda',
    area: createAirbaseAreaFromSvgCenter(211.004 - 24, 621.203 - 34),
  },
  {
    id: 'visby',
    infoUrl: '/map/airbase/visby',
    area: createAirbaseAreaFromSvgCenter(194, 682),
  },
  {
    id: 'goteborg',
    infoUrl: '/map/airbase/goteborg',
    area: createAirbaseAreaFromSvgCenter(32, 656),
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
    region: 'sweden',
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
