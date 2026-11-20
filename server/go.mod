module github.com/jamesread/greyvar/server

require (
	github.com/coder/websocket v1.8.12
	github.com/fsnotify/fsnotify v1.8.0
	github.com/jamesread/golure v0.0.0-20250821142658-fa4de6860090
	github.com/jamesread/greyvar/datlib v0.0.0
	github.com/jamesread/greyvar/world-generator v0.0.0-00010101000000-000000000000
	github.com/sirupsen/logrus v1.9.3
	google.golang.org/protobuf v1.36.11
	gopkg.in/yaml.v2 v2.4.0
)

replace github.com/jamesread/greyvar/datlib => ../datlib

replace github.com/jamesread/greyvar/world-generator => ../world-generator

require (
	github.com/aquilax/go-perlin v1.1.0 // indirect
	golang.org/x/sys v0.28.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

go 1.25.0
