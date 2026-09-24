# Greyvar

Greyvar is a casual-paced multiplayer co-op puzzle game in an 8-bit art style.

The canonical game stack is the **Go server** and **web client**. Legacy C++ client, boleas engine, and Java map editor live under `legacy/`.

## Repository layout

| Path | Description |
|------|-------------|
| `server/` | Go game server |
| `webclient/` | Browser client (Vite + Phaser) |
| `protocol/` | Protobuf definitions |
| `datlib/` | Shared Go library for game data |
| `dat-editor/` | Web editor for entity/tile definitions |
| `datlint/` | Linter for `server/dat/` |
| `res/` | Textures and tilesets |
| `world-generator/` | Procedural world generation tool |
| `tiled-plugins/` | Tiled editor extensions |
| `site/` | Project website (GitHub Pages) |
| `media/` | Screenshots and marketing assets |
| `legacy/client/` | Legacy C++ native client |
| `legacy/engine/` | Legacy boleas C++ engine |
| `legacy/editor/` | Legacy Java map editor |

## Quick start

```bash
# Generate protobuf code for server + webclient
make proto

# Build and run the server
make server
./server/greyvar-server

# Build the web client
cd webclient && npm install && make
```

Set `GREYVAR_DAT_DIR` and `GREYVAR_RES` when running tools outside the repo root. Defaults assume a standard monorepo checkout.

## Development

```bash
make tidy   # tidy all Go modules
make test   # run Go tests
make datlint
```

See component READMEs for more detail:

- [dat-editor/README.md](dat-editor/README.md)
- [datlint/README.md](datlint/README.md)
