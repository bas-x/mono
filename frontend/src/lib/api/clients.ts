import { parseApiConfigFromEnv } from '@/lib/api/config';
import { createHttpClient } from '@/lib/api/http/client';
import { createMockApiClients } from '@/lib/api/mock/clients';
import { createHealthServiceClient } from '@/lib/api/services/health';
import type { ApiClients, ApiConfig } from '@/lib/api/types';

function resolveConfig(overrides?: Partial<ApiConfig>): ApiConfig {
  return {
    ...parseApiConfigFromEnv(),
    ...overrides,
  };
}

export function createApiClients(overrides?: Partial<ApiConfig>): ApiClients {
  const config = resolveConfig(overrides);

  if (config.useMock) {
    return createMockApiClients();
  }

  const httpClient = createHttpClient(config);

  return {
    health: createHealthServiceClient(config, httpClient),
  };
}
