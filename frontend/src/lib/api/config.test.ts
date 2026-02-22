import { afterEach, describe, expect, it } from 'vitest';

import { parseApiConfigFromEnv } from '@/lib/api/config';

const mutableEnv = import.meta.env as unknown as Record<string, string | undefined>;

const originalEnv = {
  VITE_API_BASE_URL: mutableEnv.VITE_API_BASE_URL,
  VITE_WS_BASE_URL: mutableEnv.VITE_WS_BASE_URL,
  VITE_USE_MOCK_API: mutableEnv.VITE_USE_MOCK_API,
};

afterEach(() => {
  mutableEnv.VITE_API_BASE_URL = originalEnv.VITE_API_BASE_URL;
  mutableEnv.VITE_WS_BASE_URL = originalEnv.VITE_WS_BASE_URL;
  mutableEnv.VITE_USE_MOCK_API = originalEnv.VITE_USE_MOCK_API;
});

describe('parseApiConfigFromEnv', () => {
  it('uses defaults when env vars are missing', () => {
    mutableEnv.VITE_API_BASE_URL = undefined;
    mutableEnv.VITE_WS_BASE_URL = undefined;
    mutableEnv.VITE_USE_MOCK_API = undefined;

    expect(parseApiConfigFromEnv()).toEqual({
      apiBaseUrl: 'https://basex.shigure.joshuadematas.me',
      wsBaseUrl: 'wss://basex.shigure.joshuadematas.me',
      useMock: true,
    });
  });

  it('uses explicit env overrides', () => {
    mutableEnv.VITE_API_BASE_URL = 'https://example.com';
    mutableEnv.VITE_WS_BASE_URL = 'wss://example.com';
    mutableEnv.VITE_USE_MOCK_API = 'false';

    expect(parseApiConfigFromEnv()).toEqual({
      apiBaseUrl: 'https://example.com',
      wsBaseUrl: 'wss://example.com',
      useMock: false,
    });
  });

  it('falls back to default when boolean is invalid', () => {
    mutableEnv.VITE_USE_MOCK_API = 'invalid';

    expect(parseApiConfigFromEnv().useMock).toBe(true);
  });
});
