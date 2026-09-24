# Greyvar Dat Editor

Web-based editor for Greyvar **entity definitions** (`entdefs`) and **tile definitions** (`tiledefs`).

Replaces the entdef/tiledef portions of the legacy Java desktop editor with:

- **Go API server** — reads/writes `server/dat/` via datlib
- **Vite frontend** — list, create, edit, delete definitions with texture previews

Grid and world editing are out of scope.

## Prerequisites

- Go 1.21+
- Node.js 18+ (for Vite dev server)
- Clone [jamesread/greyvar](https://github.com/jamesread/greyvar) (monorepo)

## Run locally

Terminal 1 — API on `:8080`:

```bash
cd dat-editor
make server
# or: GREYVAR_DAT_DIR=../server/dat GREYVAR_RES=../res go run .
```

Terminal 2 — Vite on `https://localhost:5175` (proxies `/api` and `/res`):

```bash
cd dat-editor
make web
```

Open https://localhost:5175/ and accept the self-signed cert.

## Environment

| Variable | Default | Purpose |
|----------|---------|---------|
| `GREYVAR_DAT_DIR` | `../server/dat` | entdefs + tiledefs root |
| `GREYVAR_RES` | `../res` | texture files for previews |
| `-addr` | `:8080` | API listen address |

## API

| Method | Path | Description |
|--------|------|-------------|
| GET | `/api/health` | dat/res paths |
| GET/POST | `/api/entdefs` | list / create |
| GET/PUT/DELETE | `/api/entdefs/{name}` | read / update / delete |
| GET/POST | `/api/tiledefs` | list / create |
| GET/PUT/DELETE | `/api/tiledefs/{name}` | read / update / delete |
| GET | `/api/textures/entities` | PNG list |
| GET | `/api/textures/tiles` | PNG list |
| GET | `/res/*` | static textures |

## Build

```bash
make build   # greyvar-dat-editor binary + web/dist/
make tidy    # go mod tidy
```
