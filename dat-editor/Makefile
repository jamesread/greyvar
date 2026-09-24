.PHONY: server web dev build tidy

DAT_DIR ?= ../server/dat
RES_DIR ?= ../res

server:
	GREYVAR_DAT_DIR=$(DAT_DIR) GREYVAR_RES=$(RES_DIR) go run .

web:
	cd web && npm install && npm run dev

dev:
	@echo "Run 'make server' and 'make web' in separate terminals"

build:
	go build -o greyvar-dat-editor .
	cd web && npm install && npm run build

tidy:
	go mod tidy
