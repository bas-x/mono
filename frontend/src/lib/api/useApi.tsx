import { createContext, useContext, useMemo, useState, type ReactNode } from 'react';

import { createApiClients } from '@/lib/api/clients';
import { parseApiConfigFromEnv, resolveBaseUrls } from '@/lib/api/config';
import type { ApiClients, ApiConfig, ApiMode } from '@/lib/api/types';

type ApiContextValue = {
  clients: ApiClients;
  config: ApiConfig;
  setMode: (mode: ApiMode) => void;
  setUseMock: (useMock: boolean) => void;
};

const ApiContext = createContext<ApiContextValue | null>(null);

export function ApiProvider({ children }: { children: ReactNode }) {
  const envConfig = useMemo(() => parseApiConfigFromEnv(), []);
  const [mode, setMode] = useState<ApiMode>(envConfig.mode);

  const config = useMemo((): ApiConfig => {
    const { apiBaseUrl, wsBaseUrl } = resolveBaseUrls(mode, envConfig);
    return {
      ...envConfig,
      apiBaseUrl,
      wsBaseUrl,
      mode,
      useMock: mode === 'mock',
    };
  }, [envConfig, mode]);

  const clients = useMemo(() => createApiClients(config), [config]);

  const value = useMemo(() => ({
    clients,
    config,
    setMode,
    setUseMock: (useMock: boolean) => setMode(useMock ? 'mock' : 'remote'),
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
