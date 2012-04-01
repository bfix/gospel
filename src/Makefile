
PWD := $(shell pwd)
BASE := $(shell dirname $(PWD))

ifneq ($(shell basename $(PWD)), src)
$(error Not in source directory - you are here: '${PWD}')
endif

install:
	GOPATH=$(BASE) go install -v ./...

test:
	GOPATH=$(BASE) go test ./...

clean:
	GOPATH=$(BASE) go clean -i ./...
