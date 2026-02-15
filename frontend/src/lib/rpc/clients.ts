import type { DescService } from '@bufbuild/protobuf';
import { createClient, type Transport } from '@connectrpc/connect';

import { createMockClients } from '@/lib/rpc/mock';
import { parseRpcConfigFromEnv, createTransport } from '@/lib/rpc/transport';
import type { HealthServiceClient, RpcClients, RpcConfig } from '@/lib/rpc/types';

function createRealHealthClient(transport: Transport): HealthServiceClient {
  // TODO: replace this placeholder with generated service descriptors from src/gen.
  const pendingServiceDefinition = {} as DescService;
  const pendingClient = createClient(pendingServiceDefinition, transport);
  void pendingClient;

  return {
    async ping() {
      throw new Error(
        'Health service definition not added yet. Add generated descriptors under src/gen and wire createClient() in src/lib/rpc/clients.ts.',
      );
    },
  };
}

function resolveConfig(overrides?: Partial<RpcConfig>): RpcConfig {
  return {
    ...parseRpcConfigFromEnv(),
    ...overrides,
  };
}

export function createRpcClients(overrides?: Partial<RpcConfig>): RpcClients {
  const config = resolveConfig(overrides);

  if (config.useMock) {
    return createMockClients();
  }

  const transport = createTransport(config);
  return {
    health: createRealHealthClient(transport),
  };
}
