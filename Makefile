BUILDDATE := $(shell date +'%Y-%m-%d %H:%M:%S')

GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOGET=$(GOCMD) get -u
GOTEST=$(GOCMD) test

# Get the absolute path of the currently executing Makefile
makefilePath := $(abspath $(lastword $(MAKEFILE_LIST)))

# Extract the directory component from the Makefile's path
makefileDir := $(dir $(makefilePath))

# Define the BIN_DIRECTORY
BIN_DIRECTORY := $(makefileDir)bin

BUILDVERSION=Test

# NOTE: all golang executables install into /bin.

all: build

build: adapter mocklogic websocketserver

adapter:
	cd src/go/adapter/adaptermodule && $(GOBUILD) -ldflags '-X riden/adapter/adapter.VersionNumber=$(BUILDVERSION) -X "riden/adapter/adapter.BuildDate=$(BUILDDATE)"' -o $(BIN_DIRECTORY)/Adapter

mocklogic:
	cd src/go/mocklogic/mocklogicmodule && $(GOBUILD) -o $(BIN_DIRECTORY)/MockLogic

websocketserver:
	cd src/go/websocketserver/websocketservermodule && $(GOBUILD) -ldflags '-X riden/websocketserver/websocketserver.VersionNumber=$(BUILDVERSION) -X "riden/websocketserver/websocketserver.BuildDate=$(BUILDDATE)"' -o $(BIN_DIRECTORY)/WebSocketServer

test: adapter_test mocklogic_test websocketserver_test

adapter_test:
	# Adapter Test
	cd src/go/adapter/adaptermodule && $(GOTEST) -v

mocklogic_test:
	# MockLogic Test
	cd src/go/mocklogic/mocklogicmodule && $(GOTEST) -v

websocketserver_test:
	# WebSocketServer Test
	cd src/go/websocketserver/websocketservermodule && $(GOTEST) -v

dependencies:
	$(GOGET) github.com/gorilla/websocket

	# Logging
	$(GOGET) github.com/rs/zerolog
	$(GOGET) github.com/rs/zerolog/log

	# gRPC communication
	$(GOGET) google.golang.org/grpc
	$(GOGET) google.golang.org/protobuf