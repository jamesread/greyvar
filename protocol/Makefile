generate:
	buf generate

install: generate
	rm -f ../server/gen/greyvarprotocol/*
	cp -r out/go/* ../server/
	rm -rf ../webclient/js/gen
	mkdir -p ../webclient/js/gen
	cp -r out/js/* ../webclient/js/gen/

clean:
	rm -rf out/

.PHONY: generate install clean
