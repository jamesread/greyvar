# Greyvar Dat Editor

Web-based editor for Greyvar **entity definitions** (`entdefs`) and **tile definitions** (`tiledefs`).

Replaces the entdef/tiledef portions of the legacy Java desktop editor with:

- **Go API service** — Connect RPC over datlib
- **Vue frontend** — list, create, edit, delete definitions with texture previews

Grid and world editing are out of scope.

## Layout

| Directory | Purpose |
|-----------|---------|
| `service/` | Go Connect RPC API |
| `frontend/` | Vue + Vite web UI |
| `protocol/` | Protobuf definitions (`buf generate`) |

## Prerequisites

- Go 1.21+
- Node.js 18+
- [buf](https://buf.build/docs/installation) CLI
- Clone [jamesread/greyvar](https://github.com/jamesread/greyvar) (monorepo)

## Quick start

```bash
cd dat-editor
make setup
```

Terminal 1 — API on `:8080` (or `$PORT`):

```bash
make dev-service
```

Terminal 2 — Vite on `http://localhost:5175`:

```bash
make dev-frontend
```

Open http://localhost:5175/

## Configuration

`service/config.yaml`:

| Key | Default | Purpose |
|-----|---------|---------|
| `listen` | `:8080` | API listen address |
| `siteTitle` | Greyvar Dat Editor | UI title |
| `datDir` | _(empty)_ | entdefs + tiledefs root |
| `resDir` | _(empty)_ | texture files for previews |

Environment overrides:

| Variable | Purpose |
|----------|---------|
| `PORT` | listen port (takes precedence over `listen`) |
| `GREYVAR_DAT_DIR` | dat directory |
| `GREYVAR_RES` | res directory |
| `DEDITOR_CONFIG_FILE` | alternate config path |
| `-configdir` | directory containing `config.yaml` |

`make dev-service` sets `GREYVAR_DAT_DIR` and `GREYVAR_RES` from the monorepo layout by default.

## API

Connect RPC service: `deditor.v1.DatEditorService`

| RPC | Description |
|-----|-------------|
| `Init` | version, paths |
| `ListEntdefs` / `GetEntdef` / `CreateEntdef` / `UpdateEntdef` / `DeleteEntdef` | entity definitions |
| `ListTiledefs` / `GetTiledef` / `CreateTiledef` / `UpdateTiledef` / `DeleteTiledef` | tile definitions |
| `ListTextures` | PNG list (`kind`: `entities` or `tiles`) |

Also served:

- `GET /metrics` — Prometheus metrics
- `GET /res/*` — static textures

## Build

```bash
make build    # generate protos, service binary, frontend/dist
make test     # service + frontend tests
make generate # buf generate only
```
