# Context Engineering System

Each feature has a dedicated context folder:

`docs/context/<feature-name>/`

Expected files:
- `context.md` (mandatory)
- `api.md` (required if API or realtime transport is involved)
- `ui.md` (required if frontend-heavy)
- `simulation.md` (required if backend simulation logic changes)
- `demo.md` (required if demo flow/steps/metrics change)

## Rules
- These files are updated whenever the feature evolves.
- Keep docs short, structured, and implementation-oriented.
- Context docs are required before major features are implemented.
- PRs that change feature behavior should update corresponding context files.

## Purpose
- Give AI agents reliable domain and architecture context.
- Keep terminology and behavior consistent across frontend/backend.
- Preserve determinism and replay guarantees as the system scales.
