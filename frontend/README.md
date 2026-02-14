# Frontend

React + TypeScript + Vite app for _bas x_.

## Run from repo root

```bash
pnpm install
pnpm dev
```

Other root scripts:

```bash
pnpm build
pnpm lint
pnpm format
pnpm typecheck
```

## Run from frontend directory

```bash
cd frontend
pnpm install
pnpm dev
```

## Available scripts (`frontend/package.json`)

- `pnpm dev` - start Vite dev server
- `pnpm build` - typecheck references and build production bundle
- `pnpm preview` - preview production build locally
- `pnpm lint` - run ESLint
- `pnpm format` - run Prettier write
- `pnpm typecheck` - run TypeScript no-emit check

## Environment variables

Create `frontend/.env` or `frontend/.env.local` for local values.

Example placeholder:

```env
VITE_API_BASE_URL=http://localhost:8080
```
