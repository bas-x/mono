import type { ApiConfig } from '@/lib/api/types';

export class HttpClientError extends Error {
  readonly status: number;
  readonly path: string;
  readonly body: string;

  constructor(path: string, status: number, body: string) {
    super(`Request failed for ${path}: ${status} ${body}`);
    this.name = 'HttpClientError';
    this.path = path;
    this.status = status;
    this.body = body;
  }
}

export type RequestOptions = {
  method?: 'GET' | 'POST' | 'PUT' | 'PATCH' | 'DELETE';
  headers?: HeadersInit;
  signal?: AbortSignal;
  body?: BodyInit;
};

export type HttpClient = {
  requestJson<TResponse>(path: string, options?: RequestOptions): Promise<TResponse>;
  requestText(path: string, options?: RequestOptions): Promise<string>;
};

export function buildApiUrl(apiBaseUrl: string, path: string): string {
  const normalizedBaseUrl = apiBaseUrl.endsWith('/') ? apiBaseUrl.slice(0, -1) : apiBaseUrl;
  const normalizedPath = path.startsWith('/') ? path : `/${path}`;
  return `${normalizedBaseUrl}${normalizedPath}`;
}

function mergeHeaders(...headers: Array<HeadersInit | undefined>): Headers {
  const result = new Headers();
  for (const headerSet of headers) {
    if (!headerSet) {
      continue;
    }
    new Headers(headerSet).forEach((value, key) => {
      result.set(key, value);
    });
  }
  return result;
}

async function ensureOk(response: Response, path: string): Promise<void> {
  if (response.ok) {
    return;
  }

  const body = await response.text();
  throw new HttpClientError(path, response.status, body);
}

export function createHttpClient(config: Pick<ApiConfig, 'apiBaseUrl'>): HttpClient {
  return {
    async requestJson<TResponse>(path: string, options?: RequestOptions) {
      const response = await fetch(buildApiUrl(config.apiBaseUrl, path), {
        method: options?.method ?? 'GET',
        headers: mergeHeaders({ Accept: 'application/json' }, options?.headers),
        signal: options?.signal,
        body: options?.body,
      });

      await ensureOk(response, path);
      return (await response.json()) as TResponse;
    },

    async requestText(path: string, options?: RequestOptions) {
      const response = await fetch(buildApiUrl(config.apiBaseUrl, path), {
        method: options?.method ?? 'GET',
        headers: mergeHeaders({ Accept: 'text/plain' }, options?.headers),
        signal: options?.signal,
        body: options?.body,
      });

      await ensureOk(response, path);
      return response.text();
    },
  };
}
