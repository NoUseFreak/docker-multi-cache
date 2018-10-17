.SHELL = /bin/bash
APP = docker-multi-cache


default: clean test build all package

.PHONY: build
build:
	mkdir -p build
	go get -d
	go test
	go build -o build/${APP} *.go

all:
	go get github.com/mitchellh/gox
	mkdir -p build
	gox \
		-output="build/{{.OS}}_{{.Arch}}/${APP}"

package:
	$(shell rm -rf build/archive)
	$(shell rm -rf build/archive)
	$(eval UNIX_FILES := $(shell ls build | grep -v ${APP} | grep -v windows))
	$(eval WINDOWS_FILES := $(shell ls build | grep -v ${APP} | grep windows))
	@mkdir -p build/archive
	@for f in $(UNIX_FILES); do \
		echo Packaging $$f && \
		(cd $(shell pwd)/build/$$f && tar -czf ../archive/$$f.tar.gz ${APP}*); \
	done
	@for f in $(WINDOWS_FILES); do \
		echo Packaging $$f && \
		(cd $(shell pwd)/build/$$f && zip ../archive/$$f.zip ${APP}*); \
	done
	ls -lah build/archive/

clean:
	rm -rf build/

install:
	chmod +x build/${APP}
	sudo mv build/${APP} /usr/local/bin/${APP}

test:
	go run main.go docker build -t repo/name:version tests
