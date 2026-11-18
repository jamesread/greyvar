package api

import (
	"testing"

	"github.com/stretchr/testify/assert"

	deditorv1 "github.com/jamesread/greyvar/dat-editor/service/gen/deditor/v1"
)

func TestBuildEntdefUsesNameAsTitleFallback(t *testing.T) {
	def, err := buildEntdef("goblin", "", "idle", false, map[string]*deditorv1.EntityState{
		"idle": {Name: "idle"},
	})
	assert.NoError(t, err)
	assert.Equal(t, "goblin", def.Title)
	assert.Equal(t, "idle", def.InitialState)
}

func TestBuildEntdefFillsMissingStateNames(t *testing.T) {
	def, err := buildEntdef("goblin", "Goblin", "idle", true, map[string]*deditorv1.EntityState{
		"idle": {},
	})
	assert.NoError(t, err)
	assert.Equal(t, "idle", def.States["idle"].Name)
}
