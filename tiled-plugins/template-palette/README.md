# Template Palette

Tiled extension that provides a UI palette for selecting `.tx` object templates and placing them on object layers.

## Install

Copy or symlink this folder into a Tiled extensions directory:

- **Project:** next to your `.tiled-project` under `extensions/` (Greyvar: `res/extensions/`)
- **Global (Linux):** `~/.config/tiled/extensions/`

Or open **Edit → Preferences → Plugins** and use **Open** to reach the extensions folder.

Enable project extensions when Tiled prompts you.

## Usage

1. Open a map and select an **object layer**.
2. **View → Template Palette** (or `Ctrl+Shift+T`).
3. Optionally type in **Filter** (hides non-matching cells).
4. Click the **name button under a preview**.
5. Click the map to place instances (Insert Template tool, shortcut `V`).

Each cell shows the template’s tile graphic (first animation frame when animated), scaled 4× with nearest-neighbor filtering. Templates without a tile gid get a placeholder.

Templates are discovered by recursively scanning project folders. Use **Choose Folder…** to scan a specific directory instead, or **Use Project Folders** to clear that override.

## How it works

Tiled’s scripting API cannot assign template references on `MapObject` directly. This extension:

1. Opens the chosen `.tx` so the Templates dock / tool manager loads it
2. Activates the built-in **Insert Template** tool (`CreateTemplateTool`)

Placed objects are real template instances (they stay linked to the `.tx` file).

## Requirements

Tiled 1.10+ recommended (Dialog API, session settings, project folders). Insert Template tool selection via `MapEditor.tool` needs Tiled 1.12; older versions fall back to `tiled.trigger("CreateTemplateTool")`.
