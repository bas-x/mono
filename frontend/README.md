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

## API Layer

- `src/lib/api/*` is the only place where fetch and WebSocket are implemented.
- UI should consume hooks/clients from `@/lib/api` and never call `fetch`/`WebSocket` directly.
- HTTP endpoints used now: `/health`, `/ping`.
- Realtime stream endpoint used now: `/ws/simulation`.

## Environment variables

Create `frontend/.env` or `frontend/.env.local` for local values.

Supported variables:

- `VITE_API_BASE_URL` - backend HTTP base URL (default: `https://basex.shigure.joshuadematas.me`)
- `VITE_WS_BASE_URL` - backend WebSocket base URL (default: `wss://basex.shigure.joshuadematas.me`)
- `VITE_USE_MOCK_API` - `true` or `false` (default: `true`)

Mock mode (default):

```env
VITE_API_BASE_URL=https://basex.shigure.joshuadematas.me
VITE_WS_BASE_URL=wss://basex.shigure.joshuadematas.me
VITE_USE_MOCK_API=true
```

Real backend mode:

```env
VITE_API_BASE_URL=https://basex.shigure.joshuadematas.me
VITE_WS_BASE_URL=wss://basex.shigure.joshuadematas.me
VITE_USE_MOCK_API=false
```
