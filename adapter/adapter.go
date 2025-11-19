package adapter

// VersionNumber - Build-time variable
var VersionNumber string

// BuildDate - Build-time variable
var BuildDate string

// Connection types for communicating with clients
const (
	ConnectionTypeAll       string = "all" // Indicates a message that should be broadcast to all clients on all connections
	ConnectionTypeWebSocket string = "websocket"
)

// API message types
const (
	APIMessageTypeReserveTrip string = "reserveTrip"
	APIMessageTypeAck         string = "ack"
	APIMessageTypeAtDock      string = "atDock"
	APIMessageTypeOnBoat      string = "onBoat"
	APIMessageTypeOffBoat     string = "offBoat"
	APIMessageTypeBoatStatus  string = "boatStatus"
	APIMessageTypeArrived     string = "arrived"
)

// Service states
const (
	ServiceStateUnknown     int32 = 0
	ServiceStateOnTime      int32 = 1
	ServiceStateDelayed     int32 = 2
	ServiceStateUnavailable int32 = 3
)

var ServiceStateConversion = map[int32]string{
	ServiceStateUnknown:     "unknown",
	ServiceStateOnTime:      "onTime ",
	ServiceStateDelayed:     "delayed",
	ServiceStateUnavailable: "unavailable",
}

const (
	GangwayLocationFore string = "fore"
	GangwayLocationAft  string = "aft"
)

// API Messages

type ClientData struct {
	ConnName string
	ConnType string
}

func NewClientData(connName, connType string) ClientData {
	return ClientData{
		ConnName: connName,
		ConnType: connType,
	}
}

// Address
type Address struct {
	Number int32
	Street string
}

func NewAddress(number int32, street string) Address {
	return Address{
		Number: number,
		Street: street,
	}
}

// Dock
type Dock struct {
	Address Address
	Gangway string
}

func NewDock(address Address, gangway string) Dock {
	return Dock{
		Address: address,
		Gangway: gangway,
	}
}

// Boat
type Boat struct {
	BoatID int32
	Name   string
}

func NewBoat(boatID int32, name string) Boat {
	return Boat{
		BoatID: boatID,
		Name:   name,
	}
}

// ReserveTrip messages

// ReserveTripAPIMessage contains the Reserve message received from the client
type ReserveTripAPIMessage struct {
	MessageType     string // const "reserveTrip"
	AuthToken       string
	ClientID        string
	SourceDock      Dock
	DestinationDock Dock
}

func NewReserveTripAPIMessage(msgType, token, clientID string,
	sourceDock, destDock Dock) ReserveTripAPIMessage {
	return ReserveTripAPIMessage{
		MessageType:     msgType,
		AuthToken:       token,
		ClientID:        clientID,
		SourceDock:      sourceDock,
		DestinationDock: destDock,
	}
}

func (rm *ReserveTripAPIMessage) GetMessageType() string {
	return APIMessageTypeReserveTrip
}

// ReserveTripMockLogicMessage contains the ReserveTripAPIMessage and the ClientData for
// the client that sent the message. This is used to transmit the
// ReserveTripAPIMessage between the Adapter and the MockLogic
type ReserveTripMockLogicMessage struct {
	APIMessage ReserveTripAPIMessage
	Client     ClientData
}

func NewReserveTripMockLogicMessage(apiMsg ReserveTripAPIMessage,
	client ClientData) ReserveTripMockLogicMessage {
	return ReserveTripMockLogicMessage{
		APIMessage: apiMsg,
		Client:     client,
	}
}

// Ack messages

// AckAPIMessage contains the Ack message transmitted to the client
// as a reply to the client Reserve message
type AckAPIMessage struct {
	MessageType   string // const "ack"
	ClientID      string
	IsReserved    bool
	Boat          Boat
	TransactionID string
}

func NewAckAPIMessage(msgType, clientID string, isReserved bool,
	boat Boat, transactionID string) AckAPIMessage {
	return AckAPIMessage{
		MessageType:   msgType,
		ClientID:      clientID,
		IsReserved:    isReserved,
		Boat:          boat,
		TransactionID: transactionID,
	}
}

func (a *AckAPIMessage) GetMessageType() string {
	return APIMessageTypeAck
}

// AckMockLogicMessage contains the AckAPIMessage and the ClientData for
// the client that will receive the message. This is used to transmit the
// AckAPIMessage between the MockLogic and the Adapter
type AckMockLogicMessage struct {
	APIMessage AckAPIMessage
	Client     ClientData
}

func NewAckMockLogicMessage(apiMsg AckAPIMessage,
	client ClientData) AckMockLogicMessage {
	return AckMockLogicMessage{
		APIMessage: apiMsg,
		Client:     client,
	}
}

// AtDock messages

// AtDockAPIMessage contains the AtDock message received from the client
type AtDockAPIMessage struct {
	MessageType   string // const "atDock"
	ClientID      string
	Boat          Boat
	Dock          Dock
	TransactionID string
}

func NewAtDockAPIMessage(msgType, clientID string, boat Boat,
	dock Dock, transactionID string) AtDockAPIMessage {
	return AtDockAPIMessage{
		MessageType:   msgType,
		ClientID:      clientID,
		Boat:          boat,
		Dock:          dock,
		TransactionID: transactionID,
	}
}

func (ac *AtDockAPIMessage) GetMessageType() string {
	return APIMessageTypeAtDock
}

// AtDockMockLogicMessage contains the AtDockAPIMessage and the ClientData for
// the client that sent the message. This is used to transmit the
// AtDockAPIMessage between the Adapter and the MockLogic
type AtDockMockLogicMessage struct {
	APIMessage AtDockAPIMessage
	Client     ClientData
}

func NewAtDockMockLogicMessage(apiMsg AtDockAPIMessage,
	client ClientData) AtDockMockLogicMessage {
	return AtDockMockLogicMessage{
		APIMessage: apiMsg,
		Client:     client,
	}
}

// OnBoat messages

// OnBoatAPIMessage contains the OnBoat message received from the client
type OnBoatAPIMessage struct {
	MessageType   string // const "onBoat"
	ClientID      string
	Boat          Boat
	TransactionID string
}

func NewOnBoatAPIMessage(msgType, clientID string,
	boat Boat, transactionID string) OnBoatAPIMessage {
	return OnBoatAPIMessage{
		MessageType:   msgType,
		ClientID:      clientID,
		Boat:          boat,
		TransactionID: transactionID,
	}
}

func (ac *OnBoatAPIMessage) GetMessageType() string {
	return APIMessageTypeOnBoat
}

// OnBoatMockLogicMessage contains the OnBoatAPIMessage and the ClientData for
// the client that sent the message. This is used to transmit the
// OnBoatAPIMessage between the Adapter and the MockLogic
type OnBoatMockLogicMessage struct {
	APIMessage OnBoatAPIMessage
	Client     ClientData
}

func NewOnBoatMockLogicMessage(apiMsg OnBoatAPIMessage,
	client ClientData) OnBoatMockLogicMessage {
	return OnBoatMockLogicMessage{
		APIMessage: apiMsg,
		Client:     client,
	}
}

// OffBoat messages

// OffBoatAPIMessage contains the OffBoat message received from the client
type OffBoatAPIMessage struct {
	MessageType   string // const "offBoat"
	ClientID      string
	Boat          Boat
	TransactionID string
}

func NewOffBoatAPIMessage(msgType, clientID string,
	boat Boat, transactionID string) OffBoatAPIMessage {
	return OffBoatAPIMessage{
		MessageType:   msgType,
		ClientID:      clientID,
		Boat:          boat,
		TransactionID: transactionID,
	}
}

func (ac *OffBoatAPIMessage) GetMessageType() string {
	return APIMessageTypeOffBoat
}

// OffBoatMockLogicMessage contains the OffBoatAPIMessage and the ClientData for
// the client that sent the message. This is used to transmit the
// OffBoatAPIMessage between the Adapter and the MockLogic
type OffBoatMockLogicMessage struct {
	APIMessage OffBoatAPIMessage
	Client     ClientData
}

func NewOffBoatMockLogicMessage(apiMsg OffBoatAPIMessage,
	client ClientData) OffBoatMockLogicMessage {
	return OffBoatMockLogicMessage{
		APIMessage: apiMsg,
		Client:     client,
	}
}

// BoatStatus messages

// BoatStatusAPIMessage contains the BoatStatus message broadcast to all clients
type BoatStatusAPIMessage struct {
	MessageType  string // const "boatStatus"
	Boat         Boat
	ServiceState int32
	PreviousDock Dock
	CurrentDock  Dock
	NextDock     Dock
}

func NewBoatStatusAPIMessage(msgType string, boat Boat, serviceState int32,
	previousDock, currentDock, nextDock Dock) BoatStatusAPIMessage {
	return BoatStatusAPIMessage{
		MessageType:  msgType,
		Boat:         boat,
		ServiceState: serviceState,
		PreviousDock: previousDock,
		CurrentDock:  currentDock,
		NextDock:     nextDock,
	}
}

func (ac *BoatStatusAPIMessage) GetMessageType() string {
	return APIMessageTypeBoatStatus
}

// BoatStatusMockLogicMessage contains the BoatStatusAPIMessage and the ClientData.
// Since this message is broadcast to all clients, the ConnName is expected to be
// "allClients" and the ConnType is expected to be "all".  This is used to transmit the
// BoatStatusAPIMessage between the MockLogic and the Adapter
type BoatStatusMockLogicMessage struct {
	APIMessage BoatStatusAPIMessage
	Client     ClientData
}

func NewBoatStatusMockLogicMessage(apiMsg BoatStatusAPIMessage,
	client ClientData) BoatStatusMockLogicMessage {
	return BoatStatusMockLogicMessage{
		APIMessage: apiMsg,
		Client:     client,
	}
}

// Arrived messages

// ArrivedAPIMessage contains the Arrived message transmitted to the client
type ArrivedAPIMessage struct {
	MessageType   string // const "Arrived"
	ClientID      string
	Boat          Boat
	Dock          Dock
	TransactionID string
}

func NewArrivedAPIMessage(msgType string, clientID string,
	boat Boat, dock Dock, transactionID string) ArrivedAPIMessage {
	return ArrivedAPIMessage{
		MessageType:   msgType,
		ClientID:      clientID,
		Boat:          boat,
		Dock:          dock,
		TransactionID: transactionID,
	}
}

func (ac *ArrivedAPIMessage) GetMessageType() string {
	return APIMessageTypeArrived
}

// ArrivedMockLogicMessage contains the ArrivedAPIMessage and the ClientData for
// the client that will receive the message. This is used to transmit the
// ArrivedAPIMessage between the MockLogic and the Adapter
type ArrivedMockLogicMessage struct {
	APIMessage ArrivedAPIMessage
	Client     ClientData
}

func NewArrivedMockLogicMessage(apiMsg ArrivedAPIMessage,
	client ClientData) ArrivedMockLogicMessage {
	return ArrivedMockLogicMessage{
		APIMessage: apiMsg,
		Client:     client,
	}
}
