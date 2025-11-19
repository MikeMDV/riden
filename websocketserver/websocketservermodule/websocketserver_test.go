package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"riden/logger"
	wss "riden/websocketserver"
	"slices"
	"testing"
	"time"

	"github.com/rs/zerolog"

	"github.com/gorilla/websocket"
)

// Build Time variable
var TestHost string

// Common test parameters

var testMessageBytes = []byte{78, 52, 40}

func TestMain(m *testing.M) {
	SetUp()
	retCode := m.Run()
	TearDown()
	os.Exit(retCode)
}

func SetUp() {
	// Initialize logs
	logFile := "./websocketserver_unit_test.log"
	var err error
	Logger, err = logger.InitializeLogger(logFile, zerolog.LevelDebugValue)
	if err != nil {
		fmt.Println("Error opening log file:", logFile, ":", err)
		os.Exit(1)
	}

	// Insert any global data necessary for testing into the database here
}

func TearDown() {
	// Delete  any global data necessary for testing into the database here

}

func TestClientWebSocketHandlerHandshakeUpgrade(t *testing.T) {
	type testCase struct {
		name               string
		expectedDialError  bool
		expectedStatusCode int
		expectedHeaders    map[string]string
		expectedBody       string
	}

	cases := []testCase{
		{
			name:               "Client Websocket Handler Handshake Upgrade - Normal connection",
			expectedDialError:  false,
			expectedStatusCode: http.StatusSwitchingProtocols,
			expectedHeaders: map[string]string{
				"Connection": "Upgrade",
				"Upgrade":    "websocket",
			},
			expectedBody: "",
		},
		{
			name:               "Client Websocket Handler Handshake Upgrade - Second connection",
			expectedDialError:  false,
			expectedStatusCode: http.StatusSwitchingProtocols,
			expectedHeaders: map[string]string{
				"Connection": "Upgrade",
				"Upgrade":    "websocket",
			},
			expectedBody: "",
		},
	}

	clientWebSocketHandler := clientWebSocketHandler{
		upgrader: websocket.Upgrader{},
	}

	adapterWebSocketHandler := adapterWebSocketHandler{
		upgrader: websocket.Upgrader{},
	}

	server := httptest.NewServer(clientWebSocketHandler)
	defer server.Close()

	// Convert client server URL to ws protocol
	u, setupErr := url.Parse(server.URL)
	if setupErr != nil {
		t.Fatalf("Error parsing URL in test set-up: %s", setupErr.Error())
	}
	u.Scheme = "ws"

	adapterServer := httptest.NewServer(adapterWebSocketHandler)
	defer adapterServer.Close()

	// Convert adapter server URL to ws protocol
	au, setupErr := url.Parse(adapterServer.URL)
	if setupErr != nil {
		t.Fatalf("Error parsing URL in test set-up: %s", setupErr.Error())
	}
	au.Scheme = "ws"

	// Connect adapter
	aws, _, setupErr := websocket.DefaultDialer.Dial(au.String(), nil)
	if setupErr != nil {
		t.Fatalf("Error dialing adapter URL in test set-up: %s", setupErr.Error())
	}

	var clients []*websocket.Conn

	for _, testCase := range cases {
		// Connect client to the server
		ws, response, err := websocket.DefaultDialer.Dial(u.String(), nil)
		if testCase.expectedDialError == (err == nil) {
			t.Fatalf("Expectation of error was %v for test %s but received error was: %v",
				testCase.expectedDialError, testCase.name, err)
		}
		clients = append(clients, ws)

		// Check the status code is what we expect.
		if response.StatusCode != testCase.expectedStatusCode {
			t.Fatalf("Expected status code %d but received %d in test case: %s",
				testCase.expectedStatusCode, response.StatusCode, testCase.name)
		}
		// Check the response headers are what we expect
		for expectedKey, expectedValue := range testCase.expectedHeaders {
			retrievedVals, ok := response.Header[expectedKey]
			if !ok {
				t.Fatalf("Expected header key %s to be present but was not in test case: %s",
					expectedKey, testCase.name)
			}
			if !slices.Contains(retrievedVals, expectedValue) {
				t.Fatalf("Expected header value %s for header key %s, but value was not present in test case: %s",
					expectedValue, expectedKey, testCase.name)
			}
		}

		// Check the response body is what we expect
		bodyContents, err := io.ReadAll(response.Body)
		response.Body.Close()
		if err != nil {
			t.Fatalf("Error reading request body in test case %s: %s", testCase.name, err.Error())
		}
		if string(bodyContents) != testCase.expectedBody {
			t.Fatalf("Expected body = %s but received %s in test case: %s",
				testCase.expectedBody, string(bodyContents), testCase.name)
		}
	}

	// Clients send a close message with 1000 status code
	msg := websocket.FormatCloseMessage(websocket.CloseNormalClosure, "")
	for _, client := range clients {
		Logger.Info().Msgf("Test attempting to close remote connection: %s", client.LocalAddr().String())
		client.WriteControl(websocket.CloseMessage, msg, time.Now().Add(WriteControlDeadline))
	}

	// Adapter sends a close message with 1000 status code
	Logger.Info().Msgf("Test attempting to close remote connection: %s", aws.LocalAddr().String())
	aws.WriteControl(websocket.CloseMessage, msg, time.Now().Add(WriteControlDeadline))

	time.Sleep(500 * time.Millisecond)
}

func TestClientConnectionWithoutAdapter(t *testing.T) {
	type testCase struct {
		name               string
		expectedDialError  bool
		expectedStatusCode int
		expectedHeaders    map[string]string
		expectedBody       string
	}

	cases := []testCase{
		{
			name:               "Client Websocket Handler Handshake Upgrade - No adapter connection",
			expectedDialError:  true,
			expectedStatusCode: http.StatusInternalServerError,
			expectedHeaders: map[string]string{
				"Sec-Websocket-Version": "13",
			},
			expectedBody: "Internal Server Error\n",
		},
	}

	clientWebSocketHandler := clientWebSocketHandler{
		upgrader: websocket.Upgrader{},
	}

	server := httptest.NewServer(clientWebSocketHandler)
	defer server.Close()

	// Convert client server URL to ws protocol
	u, setupErr := url.Parse(server.URL)
	if setupErr != nil {
		t.Fatalf("Error parsing URL in test set-up: %s", setupErr.Error())
	}
	u.Scheme = "ws"

	var clients []*websocket.Conn

	for _, testCase := range cases {
		// Connect client to the server
		ws, response, err := websocket.DefaultDialer.Dial(u.String(), nil)
		if testCase.expectedDialError == (err == nil) {
			t.Fatalf("Expectation of error was %v for test %s but received error was: %v",
				testCase.expectedDialError, testCase.name, err)
		}
		// We are not expecting a connection, but append just in case it actually connected
		clients = append(clients, ws)

		// Check the status code is what we expect.
		if response.StatusCode != testCase.expectedStatusCode {
			t.Fatalf("Expected status code %d but received %d in test case: %s",
				testCase.expectedStatusCode, response.StatusCode, testCase.name)
		}
		// Check the response headers are what we expect
		for expectedKey, expectedValue := range testCase.expectedHeaders {
			retrievedVals, ok := response.Header[expectedKey]
			if !ok {
				t.Fatalf("Expected header key %s to be present but was not in test case: %s",
					expectedKey, testCase.name)
			}
			if !slices.Contains(retrievedVals, expectedValue) {
				t.Fatalf("Expected header value %s for header key %s, but value was not present in test case: %s",
					expectedValue, expectedKey, testCase.name)
			}
		}

		// Check the response body is what we expect
		bodyContents, err := io.ReadAll(response.Body)
		response.Body.Close()
		if err != nil {
			t.Fatalf("Error reading request body in test case %s: %s", testCase.name, err.Error())
		}
		if string(bodyContents) != testCase.expectedBody {
			t.Fatalf("Expected body = %s but received %s in test case: %s",
				testCase.expectedBody, string(bodyContents), testCase.name)
		}
	}

	// Clients send a close message with 1000 status code
	msg := websocket.FormatCloseMessage(websocket.CloseNormalClosure, "")
	for _, client := range clients {
		if client != nil {
			Logger.Info().Msgf("Test attempting to close remote connection: %s", client.LocalAddr().String())
			client.WriteControl(websocket.CloseMessage, msg, time.Now().Add(WriteControlDeadline))
		}
	}
}

func TestMultipleAdapterConnections(t *testing.T) {
	type testCase struct {
		name               string
		expectedDialError  bool
		expectedStatusCode int
		expectedHeaders    map[string]string
		expectedBody       string
	}

	cases := []testCase{
		{
			name:               "Adapter Websocket Handler Handshake Upgrade - Normal connection",
			expectedDialError:  false,
			expectedStatusCode: http.StatusSwitchingProtocols,
			expectedHeaders: map[string]string{
				"Connection": "Upgrade",
				"Upgrade":    "websocket",
			},
			expectedBody: "",
		},
		{
			name:               "Adapter Websocket Handler Handshake Upgrade - Second connection",
			expectedDialError:  true,
			expectedStatusCode: http.StatusTooManyRequests,
			expectedHeaders: map[string]string{
				"Sec-Websocket-Version": "13",
			},
			expectedBody: "Too Many Requests\n",
		},
	}

	adapterWebSocketHandler := adapterWebSocketHandler{
		upgrader: websocket.Upgrader{},
	}

	adapterServer := httptest.NewServer(adapterWebSocketHandler)
	defer adapterServer.Close()

	// Convert adapter server URL to ws protocol
	au, setupErr := url.Parse(adapterServer.URL)
	if setupErr != nil {
		t.Fatalf("Error parsing URL in test set-up: %s", setupErr.Error())
	}
	au.Scheme = "ws"

	var adapters []*websocket.Conn

	for _, testCase := range cases {
		// Connect client to the server
		ws, response, err := websocket.DefaultDialer.Dial(au.String(), nil)
		if testCase.expectedDialError == (err == nil) {
			t.Fatalf("Expectation of error was %v for test %s but received error was: %v",
				testCase.expectedDialError, testCase.name, err)
		}

		adapters = append(adapters, ws)

		// Check the status code is what we expect.
		if response.StatusCode != testCase.expectedStatusCode {
			t.Fatalf("Expected status code %d but received %d in test case: %s",
				testCase.expectedStatusCode, response.StatusCode, testCase.name)
		}
		// Check the response headers are what we expect
		for expectedKey, expectedValue := range testCase.expectedHeaders {
			retrievedVals, ok := response.Header[expectedKey]
			if !ok {
				t.Fatalf("Expected header key %s to be present but was not in test case: %s",
					expectedKey, testCase.name)
			}
			if !slices.Contains(retrievedVals, expectedValue) {
				t.Fatalf("Expected header value %s for header key %s, but value was not present in test case: %s",
					expectedValue, expectedKey, testCase.name)
			}
		}

		// Check the response body is what we expect
		bodyContents, err := io.ReadAll(response.Body)
		response.Body.Close()
		if err != nil {
			t.Fatalf("Error reading request body in test case %s: %s", testCase.name, err.Error())
		}
		if string(bodyContents) != testCase.expectedBody {
			t.Fatalf("Expected body = %s but received %s in test case: %s",
				testCase.expectedBody, string(bodyContents), testCase.name)
		}
	}

	// Clients send a close message with 1000 status code
	msg := websocket.FormatCloseMessage(websocket.CloseNormalClosure, "")
	for _, adapter := range adapters {
		if adapter != nil {
			Logger.Info().Msgf("Test attempting to close remote connection: %s", adapter.LocalAddr().String())
			adapter.WriteControl(websocket.CloseMessage, msg, time.Now().Add(WriteControlDeadline))
		}
	}

	time.Sleep(500 * time.Millisecond)
}

func TestSendingMessageFromClientToAdapter(t *testing.T) {
	type testCase struct {
		name                   string
		expectedMsgBytes       []byte
		expectedClientConnName string
	}

	// Set up mock adapter server and client server for test
	mockAdapterReceiveHandler := mockAdapterReceiveHandler{
		upgrader: websocket.Upgrader{},
		Message:  make(chan wss.AdapterMessage),
	}

	adapterServer := httptest.NewServer(mockAdapterReceiveHandler)
	defer adapterServer.Close()

	// Convert adapter server URL to ws protocol
	au, setupErr := url.Parse(adapterServer.URL)
	if setupErr != nil {
		t.Fatalf("Error parsing URL in test set-up: %s", setupErr.Error())
	}
	au.Scheme = "ws"

	// Connect adapter
	aws, _, setupErr := websocket.DefaultDialer.Dial(au.String(), nil)
	if setupErr != nil {
		t.Fatalf("Error dialing adapter URL in test set-up: %s", setupErr.Error())
	}

	clientWebSocketHandler := clientWebSocketHandler{
		upgrader: websocket.Upgrader{},
	}

	server := httptest.NewServer(clientWebSocketHandler)
	defer server.Close()

	// Convert client server URL to ws protocol
	u, setupErr := url.Parse(server.URL)
	if setupErr != nil {
		t.Fatalf("Error parsing URL in test set-up: %s", setupErr.Error())
	}
	u.Scheme = "ws"

	// Connect client to the server
	ws, _, setupErr := websocket.DefaultDialer.Dial(u.String(), nil)
	if setupErr != nil {
		t.Fatalf("Error dialing client URL in test set-up: %s", setupErr.Error())
	}

	// Create test cases
	cases := []testCase{
		{
			name:                   "Sending Message From Client To Adapter - Normal connection",
			expectedMsgBytes:       testMessageBytes,
			expectedClientConnName: ws.LocalAddr().String(),
		},
	}

	for _, testCase := range cases {
		// Send a message from the client to the adapter
		err := ws.WriteMessage(websocket.TextMessage, testMessageBytes)
		if err != nil {
			t.Fatalf("Error %s when writing message to server from client at %s",
				err.Error(), ws.LocalAddr().String())
		}

		recdMessage := <-mockAdapterReceiveHandler.Message

		if recdMessage.ClientConnName != testCase.expectedClientConnName {
			t.Fatalf("Expected client conn name %s but received %s in test case: %s",
				testCase.expectedClientConnName, recdMessage.ClientConnName, testCase.name)
		}

		if string(recdMessage.MessageBytes) != string(testCase.expectedMsgBytes) {
			t.Fatalf("Expected message bytes %s but received %s in test case: %s",
				string(testCase.expectedMsgBytes), string(recdMessage.MessageBytes), testCase.name)
		}
	}

	// Clients send a close message with 1000 status code
	msg := websocket.FormatCloseMessage(websocket.CloseNormalClosure, "")
	Logger.Info().Msgf("Test attempting to close remote connection: %s", ws.LocalAddr().String())
	ws.WriteControl(websocket.CloseMessage, msg, time.Now().Add(WriteControlDeadline))

	AdapterConn.Close <- 0

	// Adapter sends a close message with 1000 status code
	Logger.Info().Msgf("Test attempting to close remote connection: %s", aws.LocalAddr().String())
	aws.WriteControl(websocket.CloseMessage, msg, time.Now().Add(WriteControlDeadline))

	time.Sleep(500 * time.Millisecond)
}

func TestSendingMessageFromAdapterToClient(t *testing.T) {
	type testCase struct {
		name                   string
		expectedMsgBytes       []byte
		expectedClientConnName string
	}

	// Set up adapter server and mock client server for test
	adapterWebSocketHandler := adapterWebSocketHandler{
		upgrader: websocket.Upgrader{},
	}

	adapterServer := httptest.NewServer(adapterWebSocketHandler)
	defer adapterServer.Close()

	// Convert adapter server URL to ws protocol
	au, setupErr := url.Parse(adapterServer.URL)
	if setupErr != nil {
		t.Fatalf("Error parsing URL in test set-up: %s", setupErr.Error())
	}
	au.Scheme = "ws"

	// Connect adapter
	aws, _, setupErr := websocket.DefaultDialer.Dial(au.String(), nil)
	if setupErr != nil {
		t.Fatalf("Error dialing adapter URL in test set-up: %s", setupErr.Error())
	}

	mockClientReceiveHandler := mockClientReceiveHandler{
		upgrader: websocket.Upgrader{},
		Message:  make(chan wss.AdapterMessage),
	}

	server := httptest.NewServer(mockClientReceiveHandler)
	defer server.Close()

	// Convert client server URL to ws protocol
	u, setupErr := url.Parse(server.URL)
	if setupErr != nil {
		t.Fatalf("Error parsing URL in test set-up: %s", setupErr.Error())
	}
	u.Scheme = "ws"

	// Connect client to the server
	ws, _, setupErr := websocket.DefaultDialer.Dial(u.String(), nil)
	if setupErr != nil {
		t.Fatalf("Error dialing client URL in test set-up: %s", setupErr.Error())
	}

	// Create test cases
	cases := []testCase{
		{
			name:                   "Sending Message From Adapter To Client - Normal connection",
			expectedMsgBytes:       testMessageBytes,
			expectedClientConnName: ws.LocalAddr().String(),
		},
	}

	for _, testCase := range cases {
		// Send a message from the client to the adapter
		adapterMsg := wss.NewAdapterMessage(ws.LocalAddr().String(), testMessageBytes)
		b, err := json.Marshal(adapterMsg)
		if err != nil {
			t.Fatalf("Error %s when marshaling JSON in testCase %s",
				err.Error(), testCase.name)
		}

		err = aws.WriteMessage(websocket.TextMessage, b)
		if err != nil {
			t.Fatalf("Error %s when writing message to server from client at %s",
				err.Error(), ws.LocalAddr().String())
		}

		recdMessage := <-mockClientReceiveHandler.Message

		if recdMessage.ClientConnName != testCase.expectedClientConnName {
			t.Fatalf("Expected client conn name %s but received %s in test case: %s",
				testCase.expectedClientConnName, recdMessage.ClientConnName, testCase.name)
		}

		if string(recdMessage.MessageBytes) != string(testCase.expectedMsgBytes) {
			t.Fatalf("Expected message bytes %s but received %s in test case: %s",
				string(testCase.expectedMsgBytes), string(recdMessage.MessageBytes), testCase.name)
		}
	}

	clientVal, ok := safeClients.Load(ws.LocalAddr().String())
	if !ok {
		t.Fatal("Error deleting client from safeClients in test clean-up")
	}
	client := clientVal.(*Client)
	client.Close <- 0

	// Clients send a close message with 1000 status code
	msg := websocket.FormatCloseMessage(websocket.CloseNormalClosure, "")
	Logger.Info().Msgf("Test attempting to close remote connection: %s", ws.LocalAddr().String())
	ws.WriteControl(websocket.CloseMessage, msg, time.Now().Add(WriteControlDeadline))

	// Adapter sends a close message with 1000 status code
	Logger.Info().Msgf("Test attempting to close remote connection: %s", aws.LocalAddr().String())
	aws.WriteControl(websocket.CloseMessage, msg, time.Now().Add(WriteControlDeadline))

	time.Sleep(500 * time.Millisecond)
}
