.PHONY: all generate build build-service build-frontend test lint install setup dev-service dev-frontend clean

DAT_DIR ?= $(abspath ../server/dat)
RES_DIR ?= $(abspath ../res)

all: build

generate:
	cd protocol/proto && buf generate --template ../buf.gen.yaml

build: generate build-service build-frontend

build-service:
	GREYVAR_DAT_DIR=$(DAT_DIR) GREYVAR_RES=$(RES_DIR) $(MAKE) -wC service build

build-frontend:
	$(MAKE) -wC frontend build

test: generate
	$(MAKE) -wC service test
	$(MAKE) -wC frontend test

lint:
	$(MAKE) -wC service lint
	$(MAKE) -wC frontend lint

install:
	cd service && go mod download
	cd frontend && npm install

setup: install generate
	@echo "Setup complete. Run 'make dev-service' and 'make dev-frontend'."

dev-service:
	GREYVAR_DAT_DIR=$(DAT_DIR) GREYVAR_RES=$(RES_DIR) $(MAKE) -wC service run

dev-frontend:
	$(MAKE) -wC frontend run

clean:
	rm -rf service/dat-editor-service frontend/dist frontend/gen service/gen
