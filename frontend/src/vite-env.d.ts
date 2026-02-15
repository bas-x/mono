/// <reference types="vite/client" />

interface ImportMetaEnv {
  readonly VITE_API_BASE_URL?: string;
  readonly VITE_RPC_PROTOCOL?: 'connect' | 'grpc-web';
  readonly VITE_USE_MOCK_RPC?: 'true' | 'false';
}

interface ImportMeta {
  readonly env: ImportMetaEnv;
}
