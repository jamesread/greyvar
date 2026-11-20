package entdefs

type LootDrop struct {
	Item  string `json:"item" yaml:"item"`
	Count int32  `json:"count" yaml:"count"`
}

type EntityDefinition struct {
	Title        string                 `json:"title" yaml:"title,omitempty"`
	InitialState string                 `json:"initialState" yaml:"initialState"`
	States       map[string]EntityState `json:"states" yaml:"states"`
	// Solid entities block player movement (16×16 AABB at the entity origin).
	Solid bool `json:"solid,omitempty" yaml:"solid,omitempty"`
	// Pushable entities (with Solid) slide one tile when a player keeps
	// walking into them, if the tile beyond is free (e.g. a push block).
	Pushable bool `json:"pushable,omitempty" yaml:"pushable,omitempty"`
	// ItemID is the inventory item granted when this entity is picked up.
	ItemID string `json:"itemId,omitempty" yaml:"itemId,omitempty"`
	// Loot is granted when the entity is mined / interacted with. A map
	// object's `loot` property ("item[:count], ...") overrides it per instance.
	Loot []LootDrop `json:"loot,omitempty" yaml:"loot,omitempty"`
	// PassableStates are states in which a solid entity does not block
	// (e.g. a door's "open").
	PassableStates []string `json:"passableStates,omitempty" yaml:"passableStates,omitempty"`
	// Interact is what the interact key does to this entity.
	Interact *Interact `json:"interact,omitempty" yaml:"interact,omitempty"`

	// NPCs. Appearance draws the entity with the player's layered rig in a
	// fixed look (any unset layer/slot uses the rig default).
	Appearance *Look `json:"appearance,omitempty" yaml:"appearance,omitempty"`
	// Wander makes it stroll side to side around where the map placed it.
	Wander *Wander `json:"wander,omitempty" yaml:"wander,omitempty"`
	// Dialog names server/dat/dialogs/<dialog>.yml, opened by the interact key.
	Dialog string `json:"dialog,omitempty" yaml:"dialog,omitempty"`
	// InteractRange (px, Chebyshev) overrides how close a player must be to
	// interact, e.g. 40 to talk across a shop counter. 0 = the default (24).
	InteractRange int `json:"interactRange,omitempty" yaml:"interactRange,omitempty"`
}

// Look is a fixed layered-rig appearance (names from res/sprites/player).
type Look struct {
	// Rig is the size the NPC is drawn at: "adult" (default) or "child".
	Rig      string            `json:"rig,omitempty" yaml:"rig,omitempty"`
	Parts    map[string]string `json:"parts,omitempty" yaml:"parts,omitempty"`
	Palettes map[string]string `json:"palettes,omitempty" yaml:"palettes,omitempty"`
	HeldItem string            `json:"heldItem,omitempty" yaml:"heldItem,omitempty"`
}

// Wander: strolls to a random spot within RangePx left or right of home at
// SpeedPx every 100ms (timed, independent of the server frame rate), pausing
// PauseMs [min, max] between walks.
type Wander struct {
	RangePx int32   `json:"rangePx" yaml:"rangePx"`
	SpeedPx int32   `json:"speedPx,omitempty" yaml:"speedPx,omitempty"`
	PauseMs []int64 `json:"pauseMs,omitempty" yaml:"pauseMs,omitempty"`
}

// Interact describes interact-key behaviour. Map object properties tune it
// per instance: requires (item consumed to open), loot, msg, locked_msg.
type Interact struct {
	// Toggle cycles through these states, in order (e.g. a lever).
	Toggle []string `json:"toggle,omitempty" yaml:"toggle,omitempty"`
	// Open switches once to this state (e.g. a chest), optionally needing
	// and consuming the object's `requires` item and granting its `loot`.
	Open string `json:"open,omitempty" yaml:"open,omitempty"`
	// Close, with Open: interacting with the open entity switches it to this
	// state, and interacting again reopens it. `requires` and `loot` only
	// apply to the first opening (from the initial state).
	Close string `json:"close,omitempty" yaml:"close,omitempty"`
}

// IsPassableIn reports whether a solid entity lets players through in state.
func (e *EntityDefinition) IsPassableIn(state string) bool {
	for _, s := range e.PassableStates {
		if s == state {
			return true
		}
	}
	return false
}

// EntityState is a named gameplay state. Visual frames come from the tileset
// tile whose type/class matches title or title_state — not from YAML.
type EntityState struct {
	Name string `json:"name" yaml:"name,omitempty"`
}
