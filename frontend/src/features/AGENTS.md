# FRONTEND FEATURES KNOWLEDGE BASE

## OVERVIEW
Feature layer for map, simulation, timeline, status, and shared UI composition.

## STRUCTURE
```text
src/features/
├── map/         # map rendering, overlays, geometry, placement
├── simulation/  # setup sheet, controls, events, timeline components
├── status/      # API status panel
├── timeline/    # timeline-specific feature exports
└── ui/          # reusable shell/presentation primitives
```

## WHERE TO LOOK
| Task | Location | Notes |
|------|----------|-------|
| Main composition hub | `map/components/MapPanel.tsx` | Joins map, setup sheet, timeline, simulation hooks |
| Map-specific behavior | `map/` | Geometry, placement, overlays, view boxes |
| Simulation UI + hooks | `simulation/` | Setup forms, event hooks, control flow |
| Connectivity status | `status/components/ApiStatusPanel.tsx` | Owns status/ping/subscription concerns |
| Shared UI shells | `ui/components/` | Reusable presentation primitives |

## CONVENTIONS
- Export feature surfaces through feature `index.ts` files where they exist.
- Keep transport and environment wiring out of feature components; use `@/lib/api`.
- `MapPanel` is composition-heavy; avoid burying unrelated domain logic inside it.
- Simulation setup normalizes operator input before submission.
- Replay/timeline interactions are view-state concerns, not backend-truth mutations.

## ANTI-PATTERNS
- Direct API or WebSocket calls from feature components.
- Cross-feature coupling that bypasses the shared transport boundary.
- Duplicating map geometry or simulation form logic in unrelated components.
- Collapsing status/timeline behavior into generic UI primitives.

## NOTES
- The main hotspots under this folder are `map/` and `simulation/`; depth is capped here, so inspect those directories directly for detailed work.
