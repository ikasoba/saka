CGO_ENABLED=1

ifeq ($(OS),Windows_NT)
	GOOS=windows
	CC=x86_64-w64-mingw32-gcc
	CXX=x86_64-w64-mingw32-g++
else
	GOOS=
	CC=gcc
	CXX=g++
endif

VERSION=$(shell git describe --tag --abbrev=0)

build:
	@CGO_ENABLED=$(CGO_ENABLED) GOOS=$(GOOS) CC=$(CC) GXX=$(CXX) go build -ldflags "-X github.com/ikasoba/saka/manager.version=$(VERSION)" .
