# Turnaround Planner UI Context

## App Shell Navigation
- The planner uses a persistent left navigation rail in the Base X operations view.
- Navigation rail width is capped at `100px` to preserve map and timeline workspace.
- Rail branding uses `bas x` text-only logo until brand assets are finalized.

## Current Navigation Scope
- A single primary entry exists: `Simulation`.
- `Simulation` is intentionally a non-routing placeholder in the current build.
- Routing integration will replace the placeholder behavior without changing rail structure.

## UX Constraints
- Keep navigation styling clean and minimal to prioritize map/timeline decision workflows.
- Maintain keyboard focus visibility and semantic navigation landmarks (`nav`, accessible labels).

## Feature Composition
- API connectivity and environment status UI is encapsulated in `ApiStatusPanel` under `frontend/src/features/status`.
- The Base X operations page composes feature panels (`MapPanel`, `TimelinePanel`, `ApiStatusPanel`) and avoids embedding panel-specific logic.
- `ApiStatusPanel` owns ping interaction and simulation-stream event subscription for status display.
