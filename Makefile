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

build: adapt mocklogicmodule websocket

adapt:
	cd adapter/adaptermodule && $(GOBUILD) -ldflags '-X riden/adapter/adapter.VersionNumber=$(BUILDVERSION) -X "riden/adapter/adapter.BuildDate=$(BUILDDATE)"' -o $(BIN_DIRECTORY)/Adapter

mocklogicmodule:
	cd mocklogic/mocklogicmodule && $(GOBUILD) -o $(BIN_DIRECTORY)/MockLogic

websocket:
	cd websocketserver/websocketservermodule && $(GOBUILD) -ldflags '-X riden/websocketserver/websocketserver.VersionNumber=$(BUILDVERSION) -X "riden/websocketserver/websocketserver.BuildDate=$(BUILDDATE)"' -o $(BIN_DIRECTORY)/WebSocketServer

test: adapt_test mocklogicmodule_test websocket_test

adapt_test:
	# Adapter Test
	cd adapter/adaptermodule && $(GOTEST) -v

mocklogicmodule_test:
	# MockLogic Test
	cd mocklogic/mocklogicmodule && $(GOTEST) -v

websocket_test:
	# WebSocketServer Test
	cd websocketserver/websocketservermodule && $(GOTEST) -v

dependencies:
	$(GOGET) github.com/gorilla/websocket

	# Logging
	$(GOGET) github.com/rs/zerolog
	$(GOGET) github.com/rs/zerolog/log

	# gRPC communication
	$(GOGET) google.golang.org/grpc
	$(GOGET) google.golang.org/protobuf