package main

import (
	"github.com/jamesread/greyvar/world-generator/pkg/gridgen"
)

func main() {
	//gridgen.GenerateBiomeReticulationTest()
	
	gridgen.Generate(&gridgen.GenerationConfig{
		RowCount: uint32(16),
		ColCount: uint32(20),
		Seed: int64(1347),
		ReticulateSplines: true,
	});
}
