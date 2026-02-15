# Frontend

React + TypeScript + Vite app for _bas x_.

## Run from frontend directory

```bash
cd frontend
pnpm install
pnpm dev
```

## Run from repo root (without workspace)

```bash
pnpm --dir frontend install
pnpm --dir frontend dev
```

## Available scripts (`frontend/package.json`)

- `pnpm dev` - start Vite dev server
- `pnpm build` - typecheck references and build production bundle
- `pnpm preview` - preview production build locally
- `pnpm lint` - run ESLint
- `pnpm format` - run Prettier write
- `pnpm typecheck` - run TypeScript no-emit check
- `pnpm dev:full` - run frontend + backend together

## Environment variables

Create `frontend/.env` or `frontend/.env.local` for local values.

Supported variables:

- `VITE_API_BASE_URL` - backend base URL (default: `http://localhost:8080`)
- `VITE_RPC_PROTOCOL` - RPC protocol: `connect` or `grpc-web` (default: `connect`)
- `VITE_USE_MOCK_RPC` - `true` or `false` (default: `true`)

Mock mode (default, no backend required):

```env
VITE_API_BASE_URL=http://localhost:8080
VITE_RPC_PROTOCOL=connect
VITE_USE_MOCK_RPC=true
```

Point to a real backend:

```env
VITE_API_BASE_URL=http://localhost:8080
VITE_RPC_PROTOCOL=connect
VITE_USE_MOCK_RPC=false
```

Note: real mode currently has a placeholder Health client until generated service definitions are added under `src/gen`.
