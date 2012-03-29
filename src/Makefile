
PWD := $(shell pwd)
BASE := $(shell dirname $(PWD))

ifneq ($(shell basename $(PWD)), src)
$(error Not in source directory - you are here: '${PWD}')
endif

ifneq ($(strip $(GOPATH)),)
GOPKG := $(shell echo ${GOPATH} | cut -f 1 -d ':')
GOPATH := ${GOPATH}:$(BASE)
GOINST := cp -vr $(BASE)/pkg/* $(GOPKG)/pkg
GOCLEAN := 	rm -rf ${GOPKG}/pkg/${GOOS}_${GOARCH}/gospel
else
GOPATH := $(BASE)
GOPKG := $(BASE)
GOINST := 
GOCLEAN := 
endif

install:
	@echo GOPATH=$(GOPATH)
	$(MAKE) -C gospel/data install
	$(MAKE) -C gospel/logger install
	$(MAKE) -C gospel/math install
	$(MAKE) -C gospel/crypto install
	$(MAKE) -C gospel/bitcoin/ecc install
	$(MAKE) -C gospel/network install
	$(MAKE) -C gospel/parser install
	$(GOINST)

test:
	GOPATH=${BASE} $(MAKE) -C gospel/data test
	GOPATH=${BASE} $(MAKE) -C gospel/logger test
	GOPATH=${BASE} $(MAKE) -C gospel/math test
	GOPATH=${BASE} $(MAKE) -C gospel/crypto test
	GOPATH=${BASE} $(MAKE) -C gospel/bitcoin/ecc test
	GOPATH=${BASE} $(MAKE) -C gospel/network test
	GOPATH=${BASE} $(MAKE) -C gospel/parser test

clean:
	GOPATH=${BASE} $(MAKE) -C gospel/data clean
	GOPATH=${BASE} $(MAKE) -C gospel/logger clean
	GOPATH=${BASE} $(MAKE) -C gospel/math clean
	GOPATH=${BASE} $(MAKE) -C gospel/crypto clean
	GOPATH=${BASE} $(MAKE) -C gospel/bitcoin/ecc clean
	GOPATH=${BASE} $(MAKE) -C gospel/network clean
	GOPATH=${BASE} $(MAKE) -C gospel/parser clean
	rm -rf $(BASE)/pkg
	$(GOCLEAN)
