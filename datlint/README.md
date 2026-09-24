# datlint

Lints Greyvar `dat/` content used by the server.

## Checks

- **entdefs** (`dat/entdefs/*.yml`) — parse + each state has frames
- **legacy grids** (`dat/worlds/**/*.grid`) — load via datlib
- **tilesets** — collect unique `.tsx` refs from TMJ maps, load each via datlib before maps, report load/image errors and warnings
- **TMJ maps** (`dat/worlds/**/*.tmj`) — parse/load + object template existence

## Run

```bash
# from datlint/, discovers ../server/dat automatically when present
go run ./cmd/datlint/

# or set explicitly
GREYVAR_DAT_DIR=../server/dat ./greyvar-datlint
```

Resolution order:

1. `GREYVAR_DAT_DIR` / local `./dat` (via datlib)
2. `../server/dat`

The chosen path must contain `worlds/` and `entdefs/`. Otherwise datlint exits with an error.
