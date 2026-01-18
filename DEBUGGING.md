# Debugging / Profiling (web GUI)

The web GUI has **optional** runtime diagnostics endpoints (pprof + expvar). They are **disabled by default**.

## Enable debug endpoints

Set `QA_WEBGUI_DEBUG=1` when starting the web GUI:

```bash
QA_WEBGUI_DEBUG=1 go run ./cmd/webgui
```

Then open the app in your browser as usual. With debug enabled, these routes are available on the same local server:

- **pprof index**: `/debug/pprof/`
- **expvar**: `/debug/vars`

## Capture profiles

### Heap profile (memory)

```bash
go tool pprof -http=":0" "http://127.0.0.1:<port>/debug/pprof/heap"
```

### Goroutines (leaked goroutines / stuck work)

```bash
go tool pprof -http=":0" "http://127.0.0.1:<port>/debug/pprof/goroutine"
```

### CPU profile (hot paths)

```bash
go tool pprof -http=":0" "http://127.0.0.1:<port>/debug/pprof/profile?seconds=30"
```

### Trace (timeline)

```bash
curl -o trace.out "http://127.0.0.1:<port>/debug/pprof/trace?seconds=10"
go tool trace trace.out
```

## Notes / safety

- These endpoints can expose sensitive runtime information. Only enable them on **trusted, local-only** runs.
- The server binds to `127.0.0.1`, so it’s not reachable from other machines unless you change that.

