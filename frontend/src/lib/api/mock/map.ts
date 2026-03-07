import type { ApiAirbase, ApiAirbaseDetails } from '@/lib/api/types';

export const MOCK_AIRBASES: ApiAirbase[] = [
  {
    id: 'lulea',
    infoUrl: '/map/airbase/lulea',
    area: [
      { x: 214, y: 111 },
      { x: 225, y: 114 },
      { x: 223, y: 128 },
      { x: 211, y: 126 },
    ],
  },
  {
    id: 'arlanda',
    infoUrl: '/map/airbase/arlanda',
    area: [
      { x: 245, y: 340 },
      { x: 258, y: 343 },
      { x: 256, y: 357 },
      { x: 243, y: 354 },
    ],
  },
  {
    id: 'visby',
    infoUrl: '/map/airbase/visby',
    area: [
      { x: 212, y: 676 },
      { x: 223, y: 679 },
      { x: 220, y: 693 },
      { x: 209, y: 690 },
    ],
  },
  {
    id: 'goteborg',
    infoUrl: '/map/airbase/goteborg',
    area: [
      { x: 69, y: 580 },
      { x: 83, y: 585 },
      { x: 79, y: 600 },
      { x: 65, y: 596 },
    ],
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
