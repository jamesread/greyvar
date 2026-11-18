package recipes

type RecipeInput struct {
	Item  string `yaml:"item" json:"item"`
	Count int32  `yaml:"count" json:"count"`
}

type RecipeOutput struct {
	Item  string `yaml:"item" json:"item"`
	Count int32  `yaml:"count" json:"count"`
}

// RecipeDefinition describes a crafting recipe (inputs consumed, outputs granted).
type RecipeDefinition struct {
	ID      string         `yaml:"id" json:"id"`
	Title   string         `yaml:"title" json:"title"`
	Inputs  []RecipeInput  `yaml:"inputs" json:"inputs"`
	Outputs []RecipeOutput `yaml:"outputs" json:"outputs"`
}
