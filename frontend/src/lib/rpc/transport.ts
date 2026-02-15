import type { Transport } from '@connectrpc/connect';
import { createConnectTransport } from '@connectrpc/connect-web';

import type { RpcConfig, RpcProtocol } from '@/lib/rpc/types';

const DEFAULT_BASE_URL = 'http://localhost:8080';
const DEFAULT_PROTOCOL: RpcProtocol = 'connect';
const DEFAULT_USE_MOCK = true;

function parseProtocol(value: string | undefined): RpcProtocol {
  if (value === 'connect' || value === 'grpc-web') {
    return value;
  }
  return DEFAULT_PROTOCOL;
}

function parseUseMock(value: string | undefined): boolean {
  if (value === undefined) {
    return DEFAULT_USE_MOCK;
  }
  return value.toLowerCase() === 'true';
}

export function parseRpcConfigFromEnv(): RpcConfig {
  return {
    baseUrl: import.meta.env.VITE_API_BASE_URL?.trim() || DEFAULT_BASE_URL,
    protocol: parseProtocol(import.meta.env.VITE_RPC_PROTOCOL),
    useMock: parseUseMock(import.meta.env.VITE_USE_MOCK_RPC),
  };
}

export function createTransport(config: RpcConfig): Transport {
  if (config.protocol === 'connect') {
    return createConnectTransport({ baseUrl: config.baseUrl });
  }

  throw new Error(
    'RPC protocol "grpc-web" is configured, but grpc-web transport is not wired yet. TODO: add createGrpcWebTransport() in src/lib/rpc/transport.ts.',
  );
}
