.PHONY: build install clean

BINARY=gitflip
BINDIR=bin

build:
	@mkdir -p $(BINDIR)
	go build -o $(BINDIR)/$(BINARY) .

install: build
	install -m 0755 $(BINDIR)/$(BINARY) /usr/local/bin/$(BINARY)

clean:
	rm -rf $(BINDIR)
