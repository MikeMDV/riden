package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http/httptest"
	"net/url"
	"os"
	a "riden/adapter"
	"riden/logger"
	pb "riden/proto"
	wss "riden/websocketserver"
	"testing"
	"time"

	"github.com/rs/zerolog"

	"github.com/gorilla/websocket"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// Build Time variable
var TestHost string

// Common test parameters

var testClientConnectionName string = "testClientConnName"
var testToken string = "testToken"
var testClientID string = "testClient"
var testSourceGangway string = "fore"
var testSourceNumber int32 = 901
var testSourceStreet string = "testAvenue"
var testDestGangway string = "aft"
var testDestNumber int32 = 991
var testDestStreet string = "testBlvd"
var testServicState int32 = a.ServiceStateOnTime
var testBoatID int32 = 911
var testBoatName string = "testBoat"
var testTransactionID string = "917K-956B"

var testReserveTripAPIMessageBytes []byte
var testReserveTripAPIMessage a.ReserveTripAPIMessage
var testSourceAddress a.Address
var testSourceDock a.Dock
var testDestAddress a.Address
var testDestDock a.Dock
var testBoat a.Boat
var testClientData a.ClientData
var testReserveTripMockLogicMessage a.ReserveTripMockLogicMessage
var testAckAPIMessage a.AckAPIMessage
var testAckAPIMessageBytes []byte
var testAckAdapterMessage wss.AdapterMessage
var testAtDockAPIMessage a.AtDockAPIMessage
var testAtDockAPIMessageBytes []byte
var testAtDockMockLogicMessage a.AtDockMockLogicMessage
var testOnBoatAPIMessage a.OnBoatAPIMessage
var testOnBoatAPIMessageBytes []byte
var testOnBoatMockLogicMessage a.OnBoatMockLogicMessage
var testOffBoatAPIMessage a.OffBoatAPIMessage
var testOffBoatAPIMessageBytes []byte
var testOffBoatMockLogicMessage a.OffBoatMockLogicMessage
var testBoatStatusAPIMessage a.BoatStatusAPIMessage
var testBoatStatusAPIMessageBytes []byte
var testBoatStatusAdapterMessage wss.AdapterMessage
var testArrivedAPIMessage a.ArrivedAPIMessage
var testArrivedAPIMessageBytes []byte
var testArrivedAdapterMessage wss.AdapterMessage

func TestMain(m *testing.M) {
	SetUp()
	retCode := m.Run()
	TearDown()
	os.Exit(retCode)
}

func SetUp() {
	// Initialize logs
	logFile := "adapter_unit_test.log"
	var err error
	Logger, err = logger.InitializeLogger(logFile, zerolog.LevelDebugValue)
	if err != nil {
		fmt.Println("Error opening log file:", logFile, ":", err)
		os.Exit(1)
	}

	// Marshal a Reserve API message for testing
	testSourceAddress = a.NewAddress(testSourceNumber, testSourceStreet)
	testSourceDock = a.NewDock(testSourceAddress, testSourceGangway)
	testDestAddress = a.NewAddress(testDestNumber, testDestStreet)
	testDestDock = a.NewDock(testDestAddress, testDestGangway)
	testReserveTripAPIMessage = a.NewReserveTripAPIMessage(a.APIMessageTypeReserveTrip, testToken, testClientID,
		testSourceDock, testDestDock)
	b, err := json.Marshal(testReserveTripAPIMessage)
	if err != nil {
		Logger.Error().Msgf("Error marshaling Reserve API msg in test set-up: %s", err.Error())
	}
	testReserveTripAPIMessageBytes = b

	// Create a ReserveTripMockLogicMessage for testing
	testClientData = a.NewClientData(testClientConnectionName, a.ConnectionTypeWebSocket)
	testReserveTripMockLogicMessage = a.NewReserveTripMockLogicMessage(testReserveTripAPIMessage, testClientData)

	// Marshal an Ack API message for testing
	testBoat = a.NewBoat(testBoatID, testBoatName)
	testAckAPIMessage = a.NewAckAPIMessage(a.APIMessageTypeAck, testClientID, true,
		testBoat, testTransactionID)
	b, err = json.Marshal(testAckAPIMessage)
	if err != nil {
		Logger.Error().Msgf("Error marshaling Reserve API msg in test set-up: %s", err.Error())
	}
	testAckAPIMessageBytes = b

	// Create an Ack AdapterMessage for testing
	testAckAdapterMessage = wss.NewAdapterMessage(testClientConnectionName,
		testAckAPIMessageBytes)

	// Marshal an AtDock API message for testing
	testAtDockAPIMessage = a.NewAtDockAPIMessage(a.APIMessageTypeAtDock,
		testClientID, testBoat, testSourceDock, testTransactionID)
	b, err = json.Marshal(testAtDockAPIMessage)
	if err != nil {
		Logger.Error().Msgf("Error marshaling Reserve API msg in test set-up: %s", err.Error())
	}
	testAtDockAPIMessageBytes = b

	// Create an AtDockMockLogicMessage for testing
	testAtDockMockLogicMessage = a.NewAtDockMockLogicMessage(testAtDockAPIMessage, testClientData)

	// Marshal an OnBoat API message for testing
	testOnBoatAPIMessage = a.NewOnBoatAPIMessage(a.APIMessageTypeOnBoat,
		testClientID, testBoat, testTransactionID)
	b, err = json.Marshal(testOnBoatAPIMessage)
	if err != nil {
		Logger.Error().Msgf("Error marshaling Reserve API msg in test set-up: %s", err.Error())
	}
	testOnBoatAPIMessageBytes = b

	// Create an OnBoatMockLogicMessage for testing
	testOnBoatMockLogicMessage = a.NewOnBoatMockLogicMessage(testOnBoatAPIMessage, testClientData)

	// Marshal an OffBoat API message for testing
	testOffBoatAPIMessage = a.NewOffBoatAPIMessage(a.APIMessageTypeOffBoat,
		testClientID, testBoat, testTransactionID)
	b, err = json.Marshal(testOffBoatAPIMessage)
	if err != nil {
		Logger.Error().Msgf("Error marshaling Reserve API msg in test set-up: %s", err.Error())
	}
	testOffBoatAPIMessageBytes = b

	// Create an OffBoatMockLogicMessage for testing
	testOffBoatMockLogicMessage = a.NewOffBoatMockLogicMessage(testOffBoatAPIMessage, testClientData)

	// Marshal an BoatStatus API message for testing
	testBoatStatusAPIMessage = a.NewBoatStatusAPIMessage(a.APIMessageTypeBoatStatus,
		testBoat, testServicState, testSourceDock, testDestDock, testDestDock)
	b, err = json.Marshal(testBoatStatusAPIMessage)
	if err != nil {
		Logger.Error().Msgf("Error marshaling Reserve API msg in test set-up: %s", err.Error())
	}
	testBoatStatusAPIMessageBytes = b

	// Create a BoatStatus AdapterMessage for testing
	testBoatStatusAdapterMessage = wss.NewAdapterMessage(wss.WSSServerAllClientsConnName,
		testBoatStatusAPIMessageBytes)

	// Marshal an Arrived API message for testing
	testArrivedAPIMessage = a.NewArrivedAPIMessage(a.APIMessageTypeArrived, testClientID,
		testBoat, testSourceDock, testTransactionID)
	b, err = json.Marshal(testArrivedAPIMessage)
	if err != nil {
		Logger.Error().Msgf("Error marshaling Reserve API msg in test set-up: %s", err.Error())
	}
	testArrivedAPIMessageBytes = b

	// Create an Arrived AdapterMessage for testing
	testArrivedAdapterMessage = wss.NewAdapterMessage(testClientConnectionName,
		testArrivedAPIMessageBytes)
}

func TearDown() {
	// Delete  any global data necessary for testing into the database here

}

func TestGetMessageFromWebSocketMsg(t *testing.T) {
	type testCase struct {
		name                string
		connName            string
		messageBytes        []byte
		expectedError       bool
		expectedMessageType string
	}

	cases := []testCase{
		{
			name:                "GetMessageFromWebSocketMsg - Correct message",
			connName:            "testConn",
			messageBytes:        []byte(`{"MessageType": "atDock","ClientID": "string","CabinID": 0}`),
			expectedError:       false,
			expectedMessageType: "atDock",
		},
	}

	for _, testCase := range cases {
		adapterMsg := wss.NewAdapterMessage(testCase.connName, testCase.messageBytes)
		translatoMsgBytes, err := json.Marshal(adapterMsg)
		if err != nil {
			t.Fatalf("Error marshaling JSON in test %s",
				testCase.name)
		}
		messageType, recdAdapterMessage, err := GetMessageFromWebSocketMsg(translatoMsgBytes)

		if testCase.expectedError == (err == nil) {
			t.Fatalf("Expectation of error was %v for test %s but received error was: %v",
				testCase.expectedError, testCase.name, err)
		}

		if messageType != MessageType(testCase.expectedMessageType) {
			t.Fatalf("Expected message type %s but received %s in test case: %s",
				testCase.expectedMessageType, messageType, testCase.name)
		}

		if !bytes.Equal(recdAdapterMessage.MessageBytes, testCase.messageBytes) {
			t.Fatalf("Expected AdapterMessage.MessageBytes %s but received %s in test case: %s",
				string(testCase.messageBytes), string(recdAdapterMessage.MessageBytes), testCase.name)
		}

		if recdAdapterMessage.ClientConnName != testCase.connName {
			t.Fatalf("Expected connection name %s but received %s in test case: %s",
				testCase.connName, recdAdapterMessage.ClientConnName, testCase.name)
		}
	}
}

func TestIsRetryCodeMatch(t *testing.T) {
	type testCase struct {
		name             string
		returnedCode     int
		retryStatusCodes []string
		expectedMatch    bool
		expectedError    bool
	}

	cases := []testCase{
		{
			name:             "IsRetryCodeMatch - Match with exact code",
			returnedCode:     500,
			retryStatusCodes: []string{"500"},
			expectedMatch:    true,
			expectedError:    false,
		},
		{
			name:             "IsRetryCodeMatch - Match one of the exact codes",
			returnedCode:     429,
			retryStatusCodes: []string{"429", "5xx"},
			expectedMatch:    true,
			expectedError:    false,
		},
		{
			name:             "IsRetryCodeMatch - Match one of the wildcard codes",
			returnedCode:     500,
			retryStatusCodes: []string{"429", "5xx"},
			expectedMatch:    true,
			expectedError:    false,
		},
		{
			name:             "IsRetryCodeMatch - Match none of the codes",
			returnedCode:     301,
			retryStatusCodes: []string{"429", "5xx"},
			expectedMatch:    false,
			expectedError:    false,
		},
	}

	for _, testCase := range cases {
		matched, err := IsRetryCodeMatch(testCase.returnedCode, testCase.retryStatusCodes)

		if testCase.expectedError == (err == nil) {
			t.Fatalf("Expectation of error was %v for test %s but received error was: %v",
				testCase.expectedError, testCase.name, err)
		}

		if matched != testCase.expectedMatch {
			t.Fatalf("Expected match = %v but received %v in test case: %s",
				testCase.expectedMatch, matched, testCase.name)
		}
	}
}

func TestWSServerWriteLoop(t *testing.T) {
	type testCase struct {
		name                   string
		expectedMsgBytes       []byte
		expectedClientConnName string
	}

	// Set up mock WebSocket server for test
	mockWebSocketReceiveHandler := mockWebSocketReceiveHandler{
		upgrader: websocket.Upgrader{},
		Message:  make(chan wss.AdapterMessage),
	}

	webSocketServer := httptest.NewServer(mockWebSocketReceiveHandler)
	defer webSocketServer.Close()

	// Convert WebSocket server URL to ws protocol
	wsu, setupErr := url.Parse(webSocketServer.URL)
	if setupErr != nil {
		t.Fatalf("Error parsing URL in test set-up: %s", setupErr.Error())
	}
	wsu.Scheme = "ws"

	// Connect WebSocket server
	ad, _, setupErr := websocket.DefaultDialer.Dial(wsu.String(), nil)
	if setupErr != nil {
		t.Fatalf("Error dialing WebSocket server URL in test set-up: %s", setupErr.Error())
	}

	WebSocketServerConn = WebSocketServerConnnection{
		Conn:             ad,
		RetryStatusCodes: []string{"500"},
	}
	WebSocketServerConn.Initialize()

	// Create test cases
	cases := []testCase{
		{
			name:                   "WSServerWriteLoop - Normal connection",
			expectedMsgBytes:       testReserveTripAPIMessageBytes,
			expectedClientConnName: testClientConnectionName,
		},
	}

	for _, testCase := range cases {
		// Send a message from the Adapter to the WebSockeServer
		adapterMsg := wss.NewAdapterMessage(testClientConnectionName, testReserveTripAPIMessageBytes)
		WebSocketServerConn.Write <- adapterMsg

		recdMessage := <-mockWebSocketReceiveHandler.Message

		if recdMessage.ClientConnName != testCase.expectedClientConnName {
			t.Fatalf("Expected client conn name %s but received %s in test case: %s",
				testCase.expectedClientConnName, recdMessage.ClientConnName, testCase.name)
		}

		if string(recdMessage.MessageBytes) != string(testCase.expectedMsgBytes) {
			t.Fatalf("Expected message bytes %s but received %s in test case: %s",
				string(testCase.expectedMsgBytes), string(recdMessage.MessageBytes), testCase.name)
		}
	}
}

func TestReceivingMessageFromWebSocketServer(t *testing.T) {
	type testCase struct {
		name                   string
		expectedMessage        a.ReserveTripAPIMessage
		expectedClientConnName string
	}

	// Set up mock WebSocket server for test
	mockWebSocketReserveSendHandler := mockWebSocketReserveSendHandler{
		upgrader: websocket.Upgrader{},
	}

	webSocketServer := httptest.NewServer(mockWebSocketReserveSendHandler)
	defer webSocketServer.Close()

	// Convert WebSocket server URL to ws protocol
	wsu, setupErr := url.Parse(webSocketServer.URL)
	if setupErr != nil {
		t.Fatalf("Error parsing URL in test set-up: %s", setupErr.Error())
	}
	wsu.Scheme = "ws"

	// Connect WebSocket server
	ad, _, setupErr := websocket.DefaultDialer.Dial(wsu.String(), nil)
	if setupErr != nil {
		t.Fatalf("Error dialing WebSocket server URL in test set-up: %s", setupErr.Error())
	}

	WebSocketServerConn = WebSocketServerConnnection{
		Conn:             ad,
		RetryStatusCodes: []string{"500"},
	}
	WebSocketServerConn.Initialize()

	// Create test cases
	cases := []testCase{
		{
			name:                   "ReceivingMessageFromWebSocketServer - Normal connection",
			expectedMessage:        testReserveTripAPIMessage,
			expectedClientConnName: testClientConnectionName,
		},
	}

	// Make a new Reserve message channel in the context of this test
	GRPCChans.MakeReserveTrip()

	for _, testCase := range cases {
		time.Sleep(100 * time.Millisecond)
		// Send a ping message from the Adapter to the mock WebSockeServer
		// to trigger the mock server to send a Reserve message
		emptyMsg := []byte{}
		err := ad.WriteMessage(websocket.TextMessage, emptyMsg)
		if err != nil {
			Logger.Error().Msgf("Error %s when writing ping to mock WebSocket server", err)
		}
		Logger.Info().Msg("Wrote empty msg to mock WebSocketServer")

		reserveMockLogicMsg := <-GRPCChans.ReserveTripChannel

		if reserveMockLogicMsg.APIMessage != testCase.expectedMessage {
			t.Fatalf("Expected ReserveTripMockLogicMessage %+v but received %+v in test case: %s",
				testCase.expectedMessage, reserveMockLogicMsg.APIMessage, testCase.name)
		}

		if reserveMockLogicMsg.Client.ConnName != testCase.expectedClientConnName {
			t.Fatalf("Expected client conn name %s but received %s in test case: %s",
				testCase.expectedClientConnName, reserveMockLogicMsg.Client.ConnName, testCase.name)
		}
	}
}

func TestProcessReserveMessageToMockLogic(t *testing.T) {
	type testCase struct {
		name                   string
		expectedMessage        any
		expectedClientConnName string
	}

	// Create test cases
	cases := []testCase{
		{
			name:                   "ProcessReserveMessageToMockLogic",
			expectedMessage:        testReserveTripMockLogicMessage,
			expectedClientConnName: testClientConnectionName,
		},
	}

	// Make a new channels in the context of this test
	GRPCChans.MakeReserveTrip()

	for _, testCase := range cases {
		reserveTripToMockLogicMsg := NewMockLogicMessage(testClientConnectionName,
			a.ConnectionTypeWebSocket, a.APIMessageTypeReserveTrip, testReserveTripAPIMessageBytes)

		ProcessMessageToMockLogic(&reserveTripToMockLogicMsg)

		reserveTripMockLogicMsg := <-GRPCChans.ReserveTripChannel

		if reserveTripMockLogicMsg != testCase.expectedMessage {
			t.Fatalf("Expected ReserveTripMockLogicMessage %+v but received %+v in test case: %s",
				testCase.expectedMessage, reserveTripMockLogicMsg, testCase.name)
		}

		if reserveTripMockLogicMsg.Client.ConnName != testCase.expectedClientConnName {
			t.Fatalf("Expected client conn name %s but received %s in test case: %s",
				testCase.expectedClientConnName, reserveTripMockLogicMsg.Client.ConnName, testCase.name)
		}
	}
}

func TestProcessAtDockMessageToMockLogic(t *testing.T) {
	type testCase struct {
		name                   string
		expectedMessage        any
		expectedClientConnName string
	}

	// Create test cases
	cases := []testCase{
		{
			name:                   "ProcessAtDockMessageToMockLogic",
			expectedMessage:        testAtDockMockLogicMessage,
			expectedClientConnName: testClientConnectionName,
		},
	}

	// Make a new channels in the context of this test
	GRPCChans.MakeAtDock()

	for _, testCase := range cases {
		atDockToMockLogicMsg := NewMockLogicMessage(testClientConnectionName,
			a.ConnectionTypeWebSocket, a.APIMessageTypeAtDock, testAtDockAPIMessageBytes)

		ProcessMessageToMockLogic(&atDockToMockLogicMsg)

		atDockMockLogicMsg := <-GRPCChans.AtDockChannel

		if atDockMockLogicMsg != testCase.expectedMessage {
			t.Fatalf("Expected AtDockMockLogicMessage %+v but received %+v in test case: %s",
				testCase.expectedMessage, atDockMockLogicMsg, testCase.name)
		}

		if atDockMockLogicMsg.Client.ConnName != testCase.expectedClientConnName {
			t.Fatalf("Expected client conn name %s but received %s in test case: %s",
				testCase.expectedClientConnName, atDockMockLogicMsg.Client.ConnName, testCase.name)
		}
	}
}

func TestProcessOnBoatMessageToMockLogic(t *testing.T) {
	type testCase struct {
		name                   string
		expectedMessage        any
		expectedClientConnName string
	}

	// Create test cases
	cases := []testCase{
		{
			name:                   "ProcessOnBoatMessageToMockLogic",
			expectedMessage:        testOnBoatMockLogicMessage,
			expectedClientConnName: testClientConnectionName,
		},
	}

	// Make a new channels in the context of this test
	GRPCChans.MakeOnBoat()

	for _, testCase := range cases {
		onBoatToMockLogicMsg := NewMockLogicMessage(testClientConnectionName,
			a.ConnectionTypeWebSocket, a.APIMessageTypeOnBoat, testOnBoatAPIMessageBytes)

		ProcessMessageToMockLogic(&onBoatToMockLogicMsg)

		onBoatMockLogicMsg := <-GRPCChans.OnBoatChannel

		if onBoatMockLogicMsg != testCase.expectedMessage {
			t.Fatalf("Expected OnBoatMockLogicMessage %+v but received %+v in test case: %s",
				testCase.expectedMessage, onBoatMockLogicMsg, testCase.name)
		}

		if onBoatMockLogicMsg.Client.ConnName != testCase.expectedClientConnName {
			t.Fatalf("Expected client conn name %s but received %s in test case: %s",
				testCase.expectedClientConnName, onBoatMockLogicMsg.Client.ConnName, testCase.name)
		}
	}
}

func TestProcessOffBoatMessageToMockLogic(t *testing.T) {
	type testCase struct {
		name                   string
		expectedMessage        any
		expectedClientConnName string
	}

	// Create test cases
	cases := []testCase{
		{
			name:                   "ProcessOffBoatMessageToMockLogic",
			expectedMessage:        testOffBoatMockLogicMessage,
			expectedClientConnName: testClientConnectionName,
		},
	}

	// Make a new channels in the context of this test
	GRPCChans.MakeOffBoat()

	for _, testCase := range cases {
		offBoatToMockLogicMsg := NewMockLogicMessage(testClientConnectionName,
			a.ConnectionTypeWebSocket, a.APIMessageTypeOffBoat, testOffBoatAPIMessageBytes)

		ProcessMessageToMockLogic(&offBoatToMockLogicMsg)

		offBoatMockLogicMsg := <-GRPCChans.OffBoatChannel

		if offBoatMockLogicMsg != testCase.expectedMessage {
			t.Fatalf("Expected OffBoatMockLogicMessage %+v but received %+v in test case: %s",
				testCase.expectedMessage, offBoatMockLogicMsg, testCase.name)
		}

		if offBoatMockLogicMsg.Client.ConnName != testCase.expectedClientConnName {
			t.Fatalf("Expected client conn name %s but received %s in test case: %s",
				testCase.expectedClientConnName, offBoatMockLogicMsg.Client.ConnName, testCase.name)
		}
	}
}

func TestProcessAckMessageFromMockLogic(t *testing.T) {
	type testCase struct {
		name                   string
		expectedMessage        any
		expectedClientConnName string
	}

	// Create test cases
	cases := []testCase{
		{
			name:                   "ProcessAckMessageFromMockLogic",
			expectedMessage:        testAckAdapterMessage,
			expectedClientConnName: testClientConnectionName,
		},
	}

	// Make new channels in the context of this test
	WebSocketServerConn.Write = make(chan wss.AdapterMessage, WSChannelBufferSize)

	for _, testCase := range cases {
		ackMockLogicMsg := NewMockLogicMessage(testClientConnectionName,
			a.ConnectionTypeWebSocket, a.APIMessageTypeAck, testAckAPIMessageBytes)

		ProcessMessageFromMockLogic(&ackMockLogicMsg)

		Logger.Info().Msg("Called ProcessMessageFromMockLogic")

		ackAdapterMsg := <-WebSocketServerConn.Write

		if ackAdapterMsg.ClientConnName != testCase.expectedMessage.(wss.AdapterMessage).ClientConnName {
			t.Fatalf("Expected client conn name %s but received %s in test case: %s",
				testCase.expectedMessage.(wss.AdapterMessage).ClientConnName,
				ackAdapterMsg.ClientConnName, testCase.name)
		}

		if !bytes.Equal(ackAdapterMsg.MessageBytes, testCase.expectedMessage.(wss.AdapterMessage).MessageBytes) {
			t.Fatalf("Expected message bytes %s but received %s in test case: %s",
				string(testCase.expectedMessage.(wss.AdapterMessage).MessageBytes), string(ackAdapterMsg.MessageBytes),
				testCase.name)
		}
	}
}

func TestProcessBoatStatusMessageFromMockLogic(t *testing.T) {
	type testCase struct {
		name                   string
		expectedMessage        any
		expectedClientConnName string
	}

	// Create test cases
	cases := []testCase{
		{
			name:                   "ProcessBoatStatusMessageFromMockLogic",
			expectedMessage:        testBoatStatusAdapterMessage,
			expectedClientConnName: wss.WSSServerAllClientsConnName,
		},
	}

	// Make new channels in the context of this test
	WebSocketServerConn.Write = make(chan wss.AdapterMessage, WSChannelBufferSize)

	for _, testCase := range cases {
		boatStatusMockLogicMsg := NewMockLogicMessage("",
			a.ConnectionTypeAll, a.APIMessageTypeBoatStatus, testBoatStatusAPIMessageBytes)

		ProcessMessageFromMockLogic(&boatStatusMockLogicMsg)

		Logger.Info().Msg("Called ProcessMessageFromMockLogic")

		boatStatusAdapterMsg := <-WebSocketServerConn.Write

		if boatStatusAdapterMsg.ClientConnName != testCase.expectedClientConnName {
			t.Fatalf("Expected client conn name %s but received %s in test case: %s",
				testCase.expectedClientConnName, boatStatusAdapterMsg.ClientConnName, testCase.name)
		}

		if !bytes.Equal(boatStatusAdapterMsg.MessageBytes, testCase.expectedMessage.(wss.AdapterMessage).MessageBytes) {
			t.Fatalf("Expected message bytes %s but received %s in test case: %s",
				string(testCase.expectedMessage.(wss.AdapterMessage).MessageBytes), string(boatStatusAdapterMsg.MessageBytes),
				testCase.name)
		}
	}
}

func TestProcessArrivedMessageFromMockLogic(t *testing.T) {
	type testCase struct {
		name                   string
		expectedMessage        any
		expectedClientConnName string
	}

	// Create test cases
	cases := []testCase{
		{
			name:                   "ProcessArrivedMessageFromMockLogic",
			expectedMessage:        testArrivedAdapterMessage,
			expectedClientConnName: testClientConnectionName,
		},
	}

	// Make new channels in the context of this test
	WebSocketServerConn.Write = make(chan wss.AdapterMessage, WSChannelBufferSize)

	for _, testCase := range cases {
		arrivedMockLogicMsg := NewMockLogicMessage(testClientConnectionName,
			a.ConnectionTypeWebSocket, a.APIMessageTypeArrived, testArrivedAPIMessageBytes)

		ProcessMessageFromMockLogic(&arrivedMockLogicMsg)

		Logger.Info().Msg("Called ProcessMessageFromMockLogic")

		arrivedAdapterMsg := <-WebSocketServerConn.Write

		if arrivedAdapterMsg.ClientConnName != testCase.expectedMessage.(wss.AdapterMessage).ClientConnName {
			t.Fatalf("Expected client conn name %s but received %s in test case: %s",
				testCase.expectedMessage.(wss.AdapterMessage).ClientConnName,
				arrivedAdapterMsg.ClientConnName, testCase.name)
		}

		if !bytes.Equal(arrivedAdapterMsg.MessageBytes, testCase.expectedMessage.(wss.AdapterMessage).MessageBytes) {
			t.Fatalf("Expected message bytes %s but received %s in test case: %s",
				string(testCase.expectedMessage.(wss.AdapterMessage).MessageBytes), string(arrivedAdapterMsg.MessageBytes),
				testCase.name)
		}
	}
}

func TestSendingMessageToMockLogicWithGRPC(t *testing.T) {
	type testCase struct {
		name                   string
		expectedMessage        a.ReserveTripAPIMessage
		expectedClientConnName string
	}

	// Create test cases
	cases := []testCase{
		{
			name:                   "Sending Message To MockLogic With GRPC - ReserveTripAPIMessage",
			expectedMessage:        testReserveTripAPIMessage,
			expectedClientConnName: testClientConnectionName,
		},
	}

	// Start gRPC server in the context of this test
	go InitializeGRPCServer()

	gRPCChannel := make(chan pb.ReserveTripMessage)

	// Connect a mock gRPC client
	var opts []grpc.DialOption
	opts = append(opts, grpc.WithTransportCredentials(insecure.NewCredentials()))
	conn, err := grpc.NewClient(fmt.Sprintf("localhost:%d", 8090), opts...)
	if err != nil {
		Logger.Error().Msgf("Fail to dial gRPC: %v", err)
	}
	client := pb.NewAdapterClient(conn)
	ctx, _ := context.WithTimeout(context.Background(), 20*time.Second)
	stream, err := client.ReserveTrip(ctx)
	if err != nil {
		Logger.Error().Msgf("client.Reserve failed to create stream: %v", err)
	}

	go func() {
		for {
			in, err := stream.Recv()
			if err == io.EOF {
				// read done.
				return
			}
			if err != nil {
				Logger.Error().Msgf("client.Reserve stream.Recv() returned an error: %v", err)
				stream.CloseSend()
				return
			}
			gRPCChannel <- *in
		}
	}()

	for _, testCase := range cases {
		reserveMockLogicMsg := NewMockLogicMessage(testClientConnectionName,
			a.ConnectionTypeWebSocket, a.APIMessageTypeReserveTrip, testReserveTripAPIMessageBytes)

		ProcessMessageToMockLogic(&reserveMockLogicMsg)

		recdReserveGRPCMsg := <-gRPCChannel

		Logger.Info().Msgf("Message appeared on gRPCChannel: %+v", recdReserveGRPCMsg)

		recdSourceDockAddress := a.NewAddress(recdReserveGRPCMsg.APIMessage.SourceDock.Address.Number,
			recdReserveGRPCMsg.APIMessage.SourceDock.Address.Street)
		recdSourceDock := a.NewDock(recdSourceDockAddress,
			recdReserveGRPCMsg.APIMessage.SourceDock.Gangway)
		recdDestDockAddress := a.NewAddress(recdReserveGRPCMsg.APIMessage.DestinationDock.Address.Number,
			recdReserveGRPCMsg.APIMessage.DestinationDock.Address.Street)
		recdDestDock := a.NewDock(recdDestDockAddress,
			recdReserveGRPCMsg.APIMessage.DestinationDock.Gangway)

		recdReserveTripAPIMessage := a.NewReserveTripAPIMessage(recdReserveGRPCMsg.APIMessage.MessageType,
			recdReserveGRPCMsg.APIMessage.AuthToken, recdReserveGRPCMsg.APIMessage.ClientID,
			recdSourceDock, recdDestDock)

		if recdReserveTripAPIMessage != testCase.expectedMessage {
			t.Fatalf("Expected ReserveTripAPIMessage %+v but received %+v in test case: %s",
				testCase.expectedMessage, recdReserveTripAPIMessage, testCase.name)
		}

		if recdReserveGRPCMsg.ClientData.ConnName != testCase.expectedClientConnName {
			t.Fatalf("Expected client conn name %s but received %s in test case: %s",
				testCase.expectedClientConnName, recdReserveGRPCMsg.ClientData.ConnName, testCase.name)
		}
	}

	Logger.Info().Msg("TestSendingMessageToMockLogicWithGRPC cleaning up")
	stream.CloseSend()
	conn.Close()
	GRPCServer.Stop()
}

func TestReceivingMessageFromMockLogicWithGRPC(t *testing.T) {
	type testCase struct {
		name                   string
		expectedMessage        wss.AdapterMessage
		expectedClientConnName string
	}

	// Create test cases
	cases := []testCase{
		{
			name:                   "Receiving Message From MockLogic With GRPC - AckAPIMessage",
			expectedMessage:        testAckAdapterMessage,
			expectedClientConnName: testClientConnectionName,
		},
	}

	// Create a message for the mock gRPC client to send
	ackBoatGRPC := pb.Boat{
		BoatID: testBoatID,
		Name:   testBoatName,
	}
	ackAPIMessageGRPC := pb.AckAPIMessage{
		MessageType:   a.APIMessageTypeAck,
		ClientID:      testClientID,
		IsReserved:    true,
		Boat:          &ackBoatGRPC,
		TransactionID: testTransactionID,
	}
	ackClientDataGRPC := pb.ClientData{
		ConnName: testClientConnectionName,
		ConnType: a.ConnectionTypeWebSocket,
	}
	ackMessageGRPC := pb.AckMessage{
		APIMessage: &ackAPIMessageGRPC,
		ClientData: &ackClientDataGRPC,
	}

	// Start gRPC server in the context of this test
	go InitializeGRPCServer()

	// Connect a mock gRPC client
	var opts []grpc.DialOption
	opts = append(opts, grpc.WithTransportCredentials(insecure.NewCredentials()))
	conn, err := grpc.NewClient(fmt.Sprintf("localhost:%d", 8090), opts...)
	if err != nil {
		t.Fatalf("Fail to dial gRPC: %v", err)
	}
	defer conn.Close()
	client := pb.NewAdapterClient(conn)
	ctx, _ := context.WithTimeout(context.Background(), 20*time.Second)
	stream, err := client.Ack(ctx)
	if err != nil {
		Logger.Error().Msgf("client.Ack failed to create stream: %v", err)
	}

	go func() {
		// Wait 200 milliseconds and send the message
		time.Sleep(200 * time.Millisecond)
		if err := stream.Send(&ackMessageGRPC); err != nil {
			Logger.Error().Msgf("client.Ack: stream.Send(%+v) failed: %v", ackMessageGRPC, err)
		}
	}()

	// Make new channels in the context of this test
	WebSocketServerConn.Initialize()

	for _, testCase := range cases {
		ackAdapterMsg := <-WebSocketServerConn.Write

		Logger.Info().Msgf("Message appeared on WebSocketServerConn.Write: %+v", ackAdapterMsg)

		if ackAdapterMsg.ClientConnName != testCase.expectedMessage.ClientConnName {
			t.Fatalf("Expected client conn name %s but received %s in test case: %s",
				testCase.expectedMessage.ClientConnName,
				ackAdapterMsg.ClientConnName, testCase.name)
		}

		if !bytes.Equal(ackAdapterMsg.MessageBytes, testCase.expectedMessage.MessageBytes) {
			t.Fatalf("Expected message bytes %s but received %s in test case: %s",
				string(testCase.expectedMessage.MessageBytes), string(ackAdapterMsg.MessageBytes),
				testCase.name)
		}
	}

	Logger.Info().Msg("TestReceivingMessageFromMockLogicWithGRPC cleaning up")
	stream.CloseSend()
	conn.Close()
	GRPCServer.Stop()
}
