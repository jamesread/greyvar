package recipes

import (
	"os"
	"path/filepath"
	"strings"

	datlib "github.com/jamesread/greyvar/datlib/common"
	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"
)

func RecipesDir() string {
	return filepath.Join(datlib.DatDir(), "recipes")
}

func ReadRecipe(name string) (*RecipeDefinition, error) {
	return ReadRecipeFile(filepath.Join(RecipesDir(), name+".yml"))
}

func ReadRecipeFile(filename string) (*RecipeDefinition, error) {
	file, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	recipe := &RecipeDefinition{}
	if err := yaml.UnmarshalStrict(file, recipe); err != nil {
		log.Warnf("Cannot unmarshal recipe %v: %v", filename, err)
		return nil, err
	}

	if recipe.ID == "" {
		base := strings.TrimSuffix(filepath.Base(filename), filepath.Ext(filename))
		recipe.ID = base
	}

	return recipe, nil
}

func LoadAllRecipes() (map[string]*RecipeDefinition, error) {
	dir := RecipesDir()
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return map[string]*RecipeDefinition{}, nil
		}
		return nil, err
	}

	out := make(map[string]*RecipeDefinition)
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".yml") {
			continue
		}
		name := strings.TrimSuffix(entry.Name(), ".yml")
		recipe, err := ReadRecipe(name)
		if err != nil {
			log.Warnf("Skipping recipe %q: %v", name, err)
			continue
		}
		out[recipe.ID] = recipe
	}

	return out, nil
}
