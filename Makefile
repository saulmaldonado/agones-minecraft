.PHONY: build

all:
	$(MAKE) -C ./controller

go-build:
	$(MAKE) -C controller go-build

go-run:
	$(MAKE) -C controller go-run
