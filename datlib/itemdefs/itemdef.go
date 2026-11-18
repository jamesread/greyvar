package itemdefs

// ItemDefinition describes a stackable inventory item (separate from world entities).
type ItemDefinition struct {
	ID       string `yaml:"id" json:"id"`
	Title    string `yaml:"title" json:"title"`
	StackMax int32  `yaml:"stackMax" json:"stackMax"`
	Category string `yaml:"category" json:"category"`
}
