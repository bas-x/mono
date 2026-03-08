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
- Mock airbase positions use curated SVG-space anchors aligned to `sweden.svg`; they are not derived from a linear geo projection.
- Airbase interactions include hover inspection, click selection, and keyboard activation.
- Map region fill/stroke and airbase interaction colors are driven by theme variables in `frontend/src/styles/app.css`.
- Airbases are rendered as triangle markers positioned at base centroids, with deterministic small/medium/large capacity sizing for consistent demos.

## Data & UX Behavior

- Map overlay data is loaded via API abstraction (`src/lib/api/*`) and not fetched directly in UI components.
- V1 defaults to mock-first data source to support local demo reliability.
- Hover details are debounced and cached (TTL) with in-flight deduplication to limit redundant requests.
- Selection behavior is callback-first; map feature does not own decision panel state in V1.
- `MapPanel` is the full-screen workspace shell for the map view and controls selection state passed to `ConstellationMap`.
- The right sidebar includes mode toggles (`Live`, `Simulate`) and mode-specific action buttons.
- Mode switching is presentation-only in V1, but it must visibly change map theme colors so operators can distinguish contexts at a glance.
- In `Live` mode, selected-airbase details open in a left-side drawer triggered from the sidebar; the drawer is disabled in `Simulate`.
- In `Simulate` mode, the sidebar `Create` action opens a bottom sheet for simulation setup rather than opening another inline sidebar state.
- Control surfaces such as the navbar, sidebar, drawers, and list states use shared zinc/amber design tokens to avoid palette drift between light and dark modes.
- In `Live` mode, sidebar actions support resetting to the full-map view and selecting a base from a scrollable list.
- Selecting a base from the sidebar list must both mark it as selected in the map UI and move the map to a moderate base-focused zoom.

## Accessibility Rules

- Each airbase polygon must be keyboard reachable (`tabIndex=0`) and actionable (`Enter`/`Space`).
- Visual focus and selected state must be distinguishable from default map styling.
- Hover-only information must remain accessible via keyboard focus path.
