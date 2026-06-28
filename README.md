# Traggo MCP Server

An [MCP](https://modelcontextprotocol.io) server that lets AI assistants track time in
[Traggo](https://traggo.net/). It proxies MCP tool calls to a Traggo instance's GraphQL API,
and ships a small web UI for obtaining the access token you authenticate with.

## How it works

The server has two surfaces:

1. **Web login (`/login`)** — you sign in with your Traggo username and password and get back a
   Traggo *device token*. This is a thin wrapper over Traggo's `login` GraphQL mutation; the token
   is shown on screen for you to copy.
2. **MCP endpoint (`/mcp`)** — your MCP client connects here and sends the token as a
   `Authorization: Bearer <token>` header. Each tool call is translated into a Traggo GraphQL
   request (sent upstream as `Authorization: traggo <token>`, the scheme Traggo expects).

The server is stateless — it stores nothing. The Traggo token is the entire auth state.

```
            ┌─────────────┐   Bearer <token>   ┌──────────────┐   traggo <token>   ┌────────────┐
 MCP client │  POST /mcp  │ ─────────────────► │ traggo-mcp   │ ─────────────────► │   Traggo   │
            └─────────────┘                    │ (this server)│   GraphQL          │  instance  │
                                               └──────────────┘                    └────────────┘
   browser  ── GET/POST /login ──────────────► returns a device token to copy
```

## Tools

| Tool | Description |
|------|-------------|
| `start_timer` | Start a running timer (open timespan) with tags and a note. Start defaults to now. |
| `stop_timer` | Stop a running timer by id. End defaults to now. |
| `add_timespan` | Record a completed timespan with explicit start/end, tags and a note. |
| `update_timespan` | Edit a timespan's start, end, tags and note by id. |
| `remove_timespan` | Delete a timespan by id. |
| `list_timers` | List currently running timers (to find an id to stop). |
| `list_timespans` | List recorded timespans, optionally within an RFC3339 time range (to find an id to edit). |
| `list_tags` | List all tag definitions (key, color, usage count). |
| `create_tag` | Create a tag definition (key + hex color). |
| `update_tag` | Change a tag's color and optionally rename it. |
| `remove_tag` | Delete a tag definition by key. |
| `ping_auth` | Diagnostic — echoes the token to verify auth wiring. |

Tags are key/value pairs (e.g. `{"key": "project", "value": "traggo"}`). Times are RFC3339
(e.g. `2026-06-28T09:00:00Z`).

## Running

The server needs one environment variable: `TRAGGO_URL`, the base URL of your Traggo instance
(no trailing slash).

```sh
export TRAGGO_URL=https://traggo.example.com
go run ./cmd/server
```

It listens on `:8080`. Open <http://localhost:8080/login> to get a token, then point your MCP
client at `http://localhost:8080/mcp`.

### Connecting an MCP client

Use a streamable-HTTP MCP transport pointed at `/mcp`, with the token as a bearer credential. For
example, in a client that supports HTTP MCP servers:

```json
{
  "mcpServers": {
    "traggo": {
      "url": "http://localhost:8080/mcp",
      "headers": { "Authorization": "Bearer <your-traggo-token>" }
    }
  }
}
```

## Development

### Prerequisites

- [Go](https://go.dev/) 1.26+
- [templ](https://templ.guide/) — `go install github.com/a-h/templ/cmd/templ@latest` (HTML templating)
- [Bun](https://bun.sh/) — runs the Tailwind CLI via `bunx`
- [air](https://github.com/air-verse/air) — `go install github.com/air-verse/air@latest` (hot reload, optional)
- [Dagger](https://dagger.io/) — only needed to build/publish the container image

### Hot-reload dev loop

```sh
export TRAGGO_URL=https://traggo.example.com
air
```

`air` (configured in `.air.toml`) regenerates templates and CSS, rebuilds, and serves with live
reload on <http://localhost:8090> (proxying the app on `:8080`).

### Building assets manually

If you're not using `air`:

```sh
templ generate                                          # .templ → _templ.go
bunx @tailwindcss/cli -i ./app.css -o assets/output.css # build CSS
go run ./cmd/server
```

### Project layout

```
cmd/server/        Entry point: route wiring, starts the HTTP server
internal/
  auth/            Web login flow → Traggo device token
  mcp/             MCP server: tool registration + bearer-token middleware
  traggo/          GraphQL client for the Traggo API (all timer/timespan/tag ops)
views/             templ components and pages for the web UI
assets/            Static files (generated CSS, images)
.dagger/           Dagger module for building & publishing the container image
```

## Building & publishing the image

The container image is built with [Dagger](https://dagger.io/) — a multi-stage build (Go compile
on `golang:1.26`, runtime on `wolfi-base`) that publishes to a container registry. Registry,
credentials and image name are all parameters:

```sh
dagger call build \
  --source . \
  --registry registry.example.com \
  --username you@example.com \
  --registry-password env:REGISTRY_PASSWORD \
  --image traggo-mcp:latest
```

`--source` defaults to `.` and `--image` defaults to `traggo-mcp:latest`. The function returns the
published image reference.

## Contributing

Contributions are welcome. To get started:

1. Fork and clone the repo.
2. Install the prerequisites above and run `air` (or the manual build steps) against a Traggo
   instance you control.
3. Make your change. If you touch a `.templ` file, run `templ generate`. If you add or change a
   dependency, run `go mod tidy`.
4. Make sure `go build ./...` and `go vet ./...` pass.
5. Open a pull request describing the change.

When adding a new tool, the pattern is:

- Define the input/output types in `internal/traggo/types.go` (use `jsonschema` struct tags so the
  MCP schema is descriptive).
- Add the operation method to `internal/traggo/service.go` (it goes through the shared `execute`
  helper for auth and error handling).
- Register the tool in `internal/mcp/server.go`.

The Traggo GraphQL schema is the source of truth for available operations and field names.
