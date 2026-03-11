import { createContext, useContext, useMemo, useState, type ReactNode } from 'react';

import { createApiClients } from '@/lib/api/clients';
import { parseApiConfigFromEnv } from '@/lib/api/config';
import type { ApiClients, ApiConfig } from '@/lib/api/types';

type ApiContextValue = {
  clients: ApiClients;
  config: ApiConfig;
  setUseMock: (useMock: boolean) => void;
};

const ApiContext = createContext<ApiContextValue | null>(null);

export function ApiProvider({ children }: { children: ReactNode }) {
  const envConfig = useMemo(() => parseApiConfigFromEnv(), []);
  const [useMock, setUseMock] = useState(envConfig.useMock);

  const config = useMemo((): ApiConfig => ({
    ...envConfig,
    useMock,
  }), [envConfig, useMock]);

  const clients = useMemo(() => createApiClients(config), [config]);

  const value = useMemo(() => ({
    clients,
    config,
    setUseMock,
  }), [clients, config]);

  return (
    <ApiContext.Provider value={value}>
      {children}
    </ApiContext.Provider>
  );
}

export function useApi(): ApiContextValue {
  const context = useContext(ApiContext);
  if (!context) {
    throw new Error('useApi must be used within an ApiProvider');
  }
  return context;
}
