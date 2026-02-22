import type { HttpClient } from '@/lib/api/http/client';
import type { ApiConfig, HealthServiceClient } from '@/lib/api/types';

type HealthEndpointResponse = {
  status?: string;
};

export function createHealthServiceClient(
  _config: Pick<ApiConfig, 'apiBaseUrl'>,
  httpClient: HttpClient,
): HealthServiceClient {
  return {
    async ping(signal?: AbortSignal) {
      const [health, ping] = await Promise.all([
        httpClient.requestJson<HealthEndpointResponse>('/health', { signal }),
        httpClient.requestText('/ping', { signal }),
      ]);

      const healthStatus = health.status?.toLowerCase() === 'ok';
      const pingStatus = ping.trim().toLowerCase() === 'pong';

      return {
        ok: healthStatus && pingStatus,
        message: `health=${health.status ?? 'unknown'}, ping=${ping.trim() || 'unknown'}`,
        time: new Date().toISOString(),
      };
    },
  };
}
