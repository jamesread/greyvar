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
