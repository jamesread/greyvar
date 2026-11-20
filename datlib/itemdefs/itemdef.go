package itemdefs

// ItemDefinition describes a stackable inventory item (separate from world entities).
type ItemDefinition struct {
	ID       string `yaml:"id" json:"id"`
	Title    string `yaml:"title" json:"title"`
	StackMax int32  `yaml:"stackMax" json:"stackMax"`
	Category string `yaml:"category" json:"category"`
	// Icon is the inventory/HUD picture: a PNG path under res/
	// (img/textures/items/icons/<id>.png, 16x16).
	Icon string `yaml:"icon,omitempty" json:"icon,omitempty"`
	// Held is the optional sprite drawn in a player's hand (see res/sprites/README.md).
	Held *HeldSprite `yaml:"held,omitempty" json:"held,omitempty"`
}

// HeldSprite is a small image drawn at the pose's hand anchor.
type HeldSprite struct {
	// Sprite is a PNG path under res/, e.g. img/textures/items/held/key.png.
	Sprite string `yaml:"sprite" json:"sprite"`
	// Grip is the pixel in Sprite that sits on the hand anchor.
	Grip [2]int32 `yaml:"grip" json:"grip"`
}
