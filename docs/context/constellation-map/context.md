# Constellation Map Context

## Visual Representation Rules
- Aircraft, pads, runways, and resources are visual entities mapped from backend state.
- Visual encoding (color, icon, status) must be deterministic and documented.

## Derived State vs Backend Truth
- Backend events/state are source of truth.
- Frontend map state is a projection for rendering only.
- Frontend must not invent simulation outcomes.

## Animation Constraints
- Animations are presentational and must never change model state.
- Animation timing should be stable and replay-safe (same replay point => same visual state).

## Performance Constraints
- Target smooth interaction at demo data sizes (pan/zoom/timeline scrubbing).
- Minimize re-renders through memoized derived selectors.
- Progressive detail is acceptable when zooming or under load.
