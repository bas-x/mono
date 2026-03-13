# API PACKAGE KNOWLEDGE BASE

## OVERVIEW
Echo-based HTTP/WebSocket transport boundary for simulation operations. Thin translation layer over `services`.

## WHERE TO LOOK
| Task | Location | Notes |
|------|----------|-------|
| Server bootstrap | `server.go` | `Run` wires logger, config, deps, middleware |
| Route inventory | `routes.go` | Single source for registered REST + WS endpoints |
| Simulation handlers | `simulation.go` | Request binding, response mapping, error translation |
| Event stream transport | `websocket.go` | Client subscription / stream loop |
| Regression coverage | `websocket_test.go` | Largest API test surface |

## CONVENTIONS
- Keep handlers thin: bind, validate, delegate, translate response/error.
- Route changes belong in `routes.go`; avoid hidden registration.
- REST/WS contracts must stay aligned with `backend/docs/context/api.md` and `backend/docs/context/frontend-live-events.md`.
- Frontend consumers are expected to come through `frontend/src/lib/api/*`.

## ANTI-PATTERNS
- Embedding business logic in handlers.
- Bypassing `SimulationService` for stateful operations.
- Changing request/response shapes without doc and frontend-type follow-through.
- WebSocket behavior that blocks on slow consumers.

## COMMANDS
```bash
go test ./api
go test ./api -run TestWebsocket
```

## NOTES
- Real-time behavior is heavily test-driven here; read `websocket_test.go` before changing stream semantics.
