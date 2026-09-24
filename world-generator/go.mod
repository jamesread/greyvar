module github.com/jamesread/greyvar/world-generator

go 1.22

require (
	github.com/aquilax/go-perlin v1.1.0
	github.com/jamesread/greyvar/datlib v0.0.0
	github.com/sirupsen/logrus v1.9.3
	gopkg.in/yaml.v3 v3.0.1
)

replace github.com/jamesread/greyvar/datlib => ../datlib

require (
	golang.org/x/sys v0.28.0 // indirect
	gopkg.in/yaml.v2 v2.4.0 // indirect
)
