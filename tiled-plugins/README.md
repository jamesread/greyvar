# tiled-plugins

Tiled Map Editor extensions for Greyvar.

## Extensions

| Extension | Description |
|-----------|-------------|
| [template-palette](template-palette/) | UI palette to select `.tx` templates for placement on object layers |

## Install

Symlink or copy an extension folder into the project extensions directory:

```bash
mkdir -p ../res/extensions
ln -sfn ../../tiled-plugins/template-palette ../res/extensions/template-palette
```

Then open `res/greyvar.tiled-project` in Tiled and allow project extensions when prompted.

Global install (any project): copy into `~/.config/tiled/extensions/`.
