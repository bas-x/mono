import { useMemo } from 'react';

import { createRpcClients } from '@/lib/rpc/clients';
import { parseRpcConfigFromEnv } from '@/lib/rpc/transport';
import type { RpcClients, RpcConfig } from '@/lib/rpc/types';

export type UseRpcResult = {
  clients: RpcClients;
  config: RpcConfig;
};

export function useRpc(): UseRpcResult {
  const clients = useMemo(() => createRpcClients(), []);
  const config = useMemo(() => parseRpcConfigFromEnv(), []);

  return {
    clients,
    config,
  };
}
