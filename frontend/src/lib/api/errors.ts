import { HttpClientError } from '@/lib/api/http/client';

export function extractErrorMessage(error: unknown): string {
  if (error instanceof HttpClientError) {
    let backendMsg = error.body;
    try {
      const parsed = JSON.parse(error.body);
      if (parsed && typeof parsed.message === 'string') {
        backendMsg = parsed.message;
      }
    } catch {
      backendMsg = error.body;
    }
    return `Error ${error.status}: ${backendMsg}`;
  }

  if (typeof error === 'object' && error !== null && 'status' in error && 'body' in error) {
    const status = (error as any).status as number;
    const bodyStr = (error as any).body as string;
    let backendMsg = bodyStr;
    try {
      const parsed = JSON.parse(bodyStr);
      if (parsed && typeof parsed.message === 'string') {
        backendMsg = parsed.message;
      }
    } catch {
      backendMsg = bodyStr;
    }
    return `Error ${status}: ${backendMsg}`;
  }
  
  if (error instanceof Error) {
    return error.message;
  }
  
  return 'An unknown error occurred';
}

export function getErrorStatus(error: unknown): number | undefined {
  if (error instanceof HttpClientError) {
    return error.status;
  }
  
  if (typeof error === 'object' && error !== null && 'status' in error) {
    return (error as any).status as number;
  }
  
  return undefined;
}
