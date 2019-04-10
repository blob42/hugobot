TARGET=hugobot

GOINSTALL := GO111MODULE=on go install -v
GOBUILD := GO111MODULE=on go build -v
PKG := hugobot

.PHONY: all build install


all: build

build:
	$(GOBUILD) -o $(TARGET)

install:
	$(GOINSTALL)






