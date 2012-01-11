
include $(GOROOT)/src/Make.inc

all:
	$(MAKE) -C gospel/crypto
	$(MAKE) -C gospel/ecc
	$(MAKE) -C gospel/math
	$(MAKE) -C gospel/network
	$(MAKE) -C gospel/parser

clean:
	$(MAKE) -C gospel/crypto clean
	$(MAKE) -C gospel/ecc clean
	$(MAKE) -C gospel/math clean
	$(MAKE) -C gospel/network clean
	$(MAKE) -C gospel/parser clean


install:
	$(MAKE) -C gospel/crypto install
	$(MAKE) -C gospel/ecc install
	$(MAKE) -C gospel/math install
	$(MAKE) -C gospel/network install
	$(MAKE) -C gospel/parser install

test:
	$(MAKE) -C gospel/crypto test
	$(MAKE) -C gospel/ecc test
	$(MAKE) -C gospel/math test
#	$(MAKE) -C gospel/network test
	$(MAKE) -C gospel/parser test

bench:
	$(MAKE) -C gospel/crypto bench
	$(MAKE) -C gospel/ecc bench
	$(MAKE) -C gospel/math bench
#	$(MAKE) -C gospel/network bench
	$(MAKE) -C gospel/parser bench

nuke:
	$(MAKE) -C gospel/crypto nuke
	$(MAKE) -C gospel/ecc nuke
	$(MAKE) -C gospel/math nuke
	$(MAKE) -C gospel/network nuke
	$(MAKE) -C gospel/parser nuke
