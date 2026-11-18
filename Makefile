.PHONY: server webclient proto datlint dat-editor tidy test sprites sprites-check

server:
	$(MAKE) -C server

webclient:
	$(MAKE) -C webclient

proto:
	$(MAKE) -C protocol install

datlint:
	$(MAKE) -C datlint

dat-editor:
	$(MAKE) -C dat-editor build

tidy:
	cd datlib && go mod tidy
	cd server && go mod tidy
	cd dat-editor && go mod tidy
	cd datlint && go mod tidy
	cd world-generator && go mod tidy
	cd sprite-generator && go mod tidy

# Regenerate layered sprite sheets, rig.json and entdef .tsx from res/sprites/<rig>.
sprites:
	go run ./sprite-generator build res/sprites/player

# Fail if a rig is invalid or its generated files are stale.
sprites-check:
	go run ./sprite-generator check res/sprites/player

test: sprites-check
	cd sprite-generator && go test ./...
	cd datlib && go test ./...
	cd server && go test ./...
