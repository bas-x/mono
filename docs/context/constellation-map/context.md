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

## Interactive Overlay (V1)
- Base map is rendered as static SVG asset (`sweden.svg`) with airbase polygons overlaid in map-space coordinates.
- Overlay uses frontend map contract data (`/map`) projected into SVG space with an optional linear transform.
- Airbase interactions include hover inspection, click selection, and keyboard activation.
- Map region fill/stroke and airbase interaction colors are driven by theme variables in `frontend/src/styles/app.css`.

## Data & UX Behavior
- Map overlay data is loaded via API abstraction (`src/lib/api/*`) and not fetched directly in UI components.
- V1 defaults to mock-first data source to support local demo reliability.
- Hover details are debounced and cached (TTL) with in-flight deduplication to limit redundant requests.
- Selection behavior is callback-first; map feature does not own decision panel state in V1.
- `MapPanel` owns a compact right-side selected-airbase detail panel and controls selection state passed to `ConstellationMap`.

## Accessibility Rules
- Each airbase polygon must be keyboard reachable (`tabIndex=0`) and actionable (`Enter`/`Space`).
- Visual focus and selected state must be distinguishable from default map styling.
- Hover-only information must remain accessible via keyboard focus path.
