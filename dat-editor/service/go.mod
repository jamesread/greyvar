module github.com/jamesread/greyvar/dat-editor/service

go 1.25.0

require (
	connectrpc.com/connect v1.21.0
	github.com/jamesread/golure v0.0.0-20260902093223-27b18778c9a5
	github.com/jamesread/greyvar/datlib v0.0.0
	github.com/knadh/koanf/parsers/yaml v1.1.1
	github.com/knadh/koanf/providers/file v1.2.1
	github.com/knadh/koanf/v2 v2.3.7
	github.com/prometheus/client_golang v1.24.1
	github.com/sirupsen/logrus v1.9.3
	github.com/stretchr/testify v1.11.1
	google.golang.org/protobuf v1.36.12
)

require (
	github.com/beorn7/perks v1.0.1 // indirect
	github.com/cespare/xxhash/v2 v2.3.0 // indirect
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/fsnotify/fsnotify v1.9.0 // indirect
	github.com/go-viper/mapstructure/v2 v2.4.0 // indirect
	github.com/knadh/koanf/maps v0.1.2 // indirect
	github.com/mitchellh/copystructure v1.2.0 // indirect
	github.com/mitchellh/reflectwalk v1.0.2 // indirect
	github.com/munnerz/goautoneg v0.0.0-20191010083416-a7dc8b61c822 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/prometheus/client_model v0.6.2 // indirect
	github.com/prometheus/common v0.70.1 // indirect
	github.com/prometheus/procfs v0.21.1 // indirect
	go.yaml.in/yaml/v3 v3.0.3 // indirect
	golang.org/x/sys v0.47.0 // indirect
	gopkg.in/yaml.v2 v2.4.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

replace github.com/jamesread/greyvar/datlib => ../../datlib
