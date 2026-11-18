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
	// Loot is granted when the entity is mined / interacted with.
	Loot []LootDrop `json:"loot,omitempty" yaml:"loot,omitempty"`
}

// EntityState is a named gameplay state. Visual frames come from the tileset
// tile whose type/class matches title or title_state — not from YAML.
type EntityState struct {
	Name string `json:"name" yaml:"name,omitempty"`
}
