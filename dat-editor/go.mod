module github.com/jamesread/greyvar/dat-editor

go 1.21

require (
	github.com/jamesread/greyvar/datlib v0.0.0
	github.com/sirupsen/logrus v1.9.3
)

require (
	golang.org/x/sys v0.28.0 // indirect
	gopkg.in/yaml.v2 v2.4.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

replace github.com/jamesread/greyvar/datlib => ../datlib
