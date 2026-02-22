import { useMemo } from 'react';

import { createApiClients } from '@/lib/api/clients';
import { parseApiConfigFromEnv } from '@/lib/api/config';
import type { ApiClients, ApiConfig } from '@/lib/api/types';

export type UseApiResult = {
  clients: ApiClients;
  config: ApiConfig;
};

export function useApi(): UseApiResult {
  const config = useMemo(() => parseApiConfigFromEnv(), []);
  const clients = useMemo(() => createApiClients(config), [config]);

  return {
    clients,
    config,
  };
}
