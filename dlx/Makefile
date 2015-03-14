include ../Makefile.inc

all: test build

get:
	$(GOGET) -d -v

install: get
	$(GOINSTALL) -v

test: install
	$(GOTEST)

build:
	$(GOBUILD)
