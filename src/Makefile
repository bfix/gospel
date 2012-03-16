
PWD := $(shell pwd)
BASE := $(shell dirname $(PWD))

ifneq ($(shell basename $(PWD)), src)
$(error Not in source directory - you are here: '${PWD}')
endif

#ifdef ${GOPATH}
#GOPATH := ${GOPATH}:$(BASE)
#else
#GOPATH := $(BASE)
#endif

all:
	$(MAKE) -C gospel/data
	$(MAKE) -C gospel/logger
	$(MAKE) -C gospel/math
	$(MAKE) -C gospel/crypto
	$(MAKE) -C gospel/ecc
	$(MAKE) -C gospel/network
	$(MAKE) -C gospel/parser

clean:
	$(MAKE) -C gospel/data clean
	$(MAKE) -C gospel/logger clean
	$(MAKE) -C gospel/math clean
	$(MAKE) -C gospel/crypto clean
	$(MAKE) -C gospel/ecc clean
	$(MAKE) -C gospel/network clean
	$(MAKE) -C gospel/parser clean
	rm -rf $(BASE)/pkg

test:
	$(MAKE) -C gospel/data test
	$(MAKE) -C gospel/logger test
	$(MAKE) -C gospel/math test
	$(MAKE) -C gospel/crypto test
	$(MAKE) -C gospel/ecc test
	$(MAKE) -C gospel/network test
	$(MAKE) -C gospel/parser test

install:
	$(MAKE) -C gospel/data install
	$(MAKE) -C gospel/logger install
	$(MAKE) -C gospel/math install
	$(MAKE) -C gospel/crypto install
	$(MAKE) -C gospel/ecc install
	$(MAKE) -C gospel/network install
	$(MAKE) -C gospel/parser install

distr:	install
	cp -vr $(BASE)/pkg/* $(GOROOT)/pkg
