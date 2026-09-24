.PHONY: server webclient proto datlint dat-editor tidy test

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

test:
	cd datlib && go test ./...
	cd server && go test ./...
