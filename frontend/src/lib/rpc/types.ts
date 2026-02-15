export type RpcProtocol = 'connect' | 'grpc-web';

export type HealthPingResult = {
  ok: boolean;
  message: string;
  time: string;
};

export type RpcConfig = {
  baseUrl: string;
  protocol: RpcProtocol;
  useMock: boolean;
};

export interface HealthServiceClient {
  ping(): Promise<HealthPingResult>;
}

export type RpcClients = {
  health: HealthServiceClient;
};
