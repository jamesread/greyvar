# NPCs

An NPC is an entity whose entdef gives it a look, a stroll and/or a dialog.
Bill (`entdefs/bill.yml`, `dialogs/bill.yml`, in the red house on isleOfStarting 1.2) is the
example.

```yaml
title: bill            # entdef name (the map object's type)
solid: true
initialState: idle
states: {idle: {}}
appearance:            # a fixed look, drawn with the adult rig (res/sprites/adult)
  parts: {hair: short, beard: beard}
  palettes: {skin: tan, hair: grey, shirt: green, pants: khaki}
wander:                # optional: strolls side to side around where placed
  rangePx: 24          # how far left/right of home
  speedPx: 2           # per 100ms
  pauseMs: [1500, 4000]
dialog: bill           # optional: dialogs/bill.yml, opened with E
```

- NPCs are adults. `appearance.rig: child` draws one at the small (child)
  size instead, with res/sprites/player's styles. datlint checks every look
  against its rig.
- Place it in Tiled as an object with `type` set to the entdef name.
- Leave room for the stroll: `rangePx` either side, on the same row.
- NPCs only move while a player is on their grid; they never walk into
  walls, solid entities or players, and stop (facing you) while you talk.
- Dialog format (nodes, options, `goto`, `give: "item[:count]"`): see
  `datlib/dialogs`. `datlint` checks every dialog and entdef `dialog:` name.

## Dialog options that change the player's look

An option's `wear` applies a look change when picked, e.g. a tailor
(Geraldine, `dialogs/geraldine.yml`, in `0.0.tailorShop.tmj`):

```yaml
- {label: "Blue", wear: {shirt: blue}, goto: done}
```

Keys are colour slots (`shirt`, `pants`, `hair`, `skin`, `garment`, `trim`)
or layers (`hair`, `beard`, `outfit`); values come from the adult rig,
`res/sprites/adult/rig.yml`, and datlint rejects any it does not have. `hair`
is both: a variant of the hair layer (`short`, `long`, `ponytail`, `bun`,
`curly`, `bob`, `bald`) changes the style, anything else (`blonde`) the
colour. Options matching what
the player already wears are shown with "(wearing)". The new look is sent to
everyone and the player's client saves it, like the `appearance` command.

Geraldine also sells dresses: the `outfit` layer, drawn over shirt and
trousers, in a `garment` colour with a `trim` colour
(`wear: {outfit: dress, garment: pink}`, then `wear: {trim: gold}`;
`outfit: none` takes it off). She wears one herself (purple with gold trim),
with her hair in a bun.
Skirts (`pants: skirt`) are no longer sold.

Professor Ivan (`dialogs/ivan.yml`, `0.0.professorHouse.tmj`, beside the
tailor's shop) uses the same mechanism for skin tone: `wear: {skin: deep}`.

Claire (`dialogs/claire.yml`, `1.2.salon.tmj`, across from Bill's house)
changes hairstyle (all seven: `wear: {hair: bob}`), hair colour (`wear: {hair:
red}`) and beard (`wear: {beard: beard}` / `{beard: none}`; a beard takes the
hair colour).

The NPCs' own looks: Bill short grey hair and a beard; Geraldine a red bun
and her dress; Ivan bald with a grey beard; Claire a blonde bob; Herman short
brown hair, dressed in green.

## Mannequins: outfits to try on

Geraldine's shop has four `mannequin`s (`entdefs/mannequin.yml`): faceless
figures (the adult rig's `form` layer) wearing real outfits. Each is a map
object with two properties:

```
wear:  outfit=dress,garment=blue,trim=white     # like a dialog option's wear
title: blue dress                               # for the message
```

Using one (E) puts the outfit on the player ("You put on the blue dress.").
Include `outfit=none` in a shirt-and-trousers outfit so it takes a dress off.
Mannequins have the grey `mannequin` skin (a hidden palette: never given to
a new player at random), so a dress's bare arms match the plaster head.
Any figure with an appearance and a `wear` property works the same way;
datlint checks the values against the adult rig. Mannequins stand still
(state `stand`, no breathing).

## Herman's Rift Bank (accounts)

Herman (`entdefs/herman.yml`, `dialogs/herman.yml`, in `0.1.hermanHouse.tmj`,
the purple-roofed house on isleOfStarting 0.1) banks rift coins. A dialog
node's `deposit: <item>` moves every one the player carries into their
account, and node text can show it with `{balance:<item>}`:

```yaml
  deposited:
    text: "All counted and locked in the vault."
    deposit: rift_coin          # appends "(Deposited Rift Coin x4. Balance: 12)"
```

Accounts outlast the inventory: the server keeps them per username in
`$GREYVAR_STATE_DIR/accounts.json`. `datlint` checks deposit/balance items.

Herman stands behind his counter, so his entdef sets `interactRange: 52`
(px, Chebyshev; default 24): players reach him from the other side.
