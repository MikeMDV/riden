package main

import (
	"encoding/json"
	"io"
	a "riden/adapter"
	pb "riden/proto"
	wss "riden/websocketserver"
)

// GRPCChannels holds the channels that will be used to pass messages to the
// MockLogic, using the gRPC bi-directional streams
type GRPCChannels struct {
	ReserveTripChannel chan a.ReserveTripMockLogicMessage
	AtDockChannel      chan a.AtDockMockLogicMessage
	OnBoatChannel      chan a.OnBoatMockLogicMessage
	OffBoatChannel     chan a.OffBoatMockLogicMessage
}

func (grpcc *GRPCChannels) MakeReserveTrip() {
	grpcc.ReserveTripChannel = make(chan a.ReserveTripMockLogicMessage, GRPCChannelBufferSize)
}

func (grpcc *GRPCChannels) CloseReserveTrip() {
	close(grpcc.ReserveTripChannel)
}

func (grpcc *GRPCChannels) MakeAtDock() {
	grpcc.AtDockChannel = make(chan a.AtDockMockLogicMessage, GRPCChannelBufferSize)
}

func (grpcc *GRPCChannels) CloseAtDock() {
	close(grpcc.AtDockChannel)
}

func (grpcc *GRPCChannels) MakeOnBoat() {
	grpcc.OnBoatChannel = make(chan a.OnBoatMockLogicMessage, GRPCChannelBufferSize)
}

func (grpcc *GRPCChannels) CloseOnBoat() {
	close(grpcc.OnBoatChannel)
}

func (grpcc *GRPCChannels) MakeOffBoat() {
	grpcc.OffBoatChannel = make(chan a.OffBoatMockLogicMessage, GRPCChannelBufferSize)
}

func (grpcc *GRPCChannels) CloseOffBoat() {
	close(grpcc.OffBoatChannel)
}

var GRPCChannelBufferSize int = 1024

var GRPCChans GRPCChannels

// adapterServer is used to implement adapter.AdapterServer
type adapterServer struct {
	pb.UnimplementedAdapterServer
}

// ProcessMessageToMockLogic processes a message that is being sent to
// the MockLogic
func ProcessMessageToMockLogic(mlMsg *MockLogicMessage) {
	switch mlMsg.MessageType {
	case a.APIMessageTypeReserveTrip:
		Logger.Info().Msgf("Processing %s message to MockLogic from ConnName: %s, ConnType: %s",
			mlMsg.MessageType, mlMsg.ConnName, mlMsg.ConnType)
		clientData := a.NewClientData(mlMsg.ConnName, mlMsg.ConnType)
		var apiMsg a.ReserveTripAPIMessage
		err := json.Unmarshal(mlMsg.APIMessageBytes, &apiMsg)
		if err != nil {
			Logger.Debug().Msgf("Error unmarshaling %s message: %s",
				mlMsg.MessageType, err.Error())
			return
		}
		reserveTripMockLogicMsg := a.NewReserveTripMockLogicMessage(apiMsg, clientData)

		GRPCChans.ReserveTripChannel <- reserveTripMockLogicMsg

	case a.APIMessageTypeAtDock:
		Logger.Info().Msgf("Processing %s message to MockLogic from ConnName: %s, ConnType: %s",
			mlMsg.MessageType, mlMsg.ConnName, mlMsg.ConnType)
		clientData := a.NewClientData(mlMsg.ConnName, mlMsg.ConnType)
		var apiMsg a.AtDockAPIMessage
		err := json.Unmarshal(mlMsg.APIMessageBytes, &apiMsg)
		if err != nil {
			Logger.Debug().Msgf("Error unmarshaling %s message: %s",
				mlMsg.MessageType, err.Error())
			return
		}

		atDockMockLogicMessage := a.NewAtDockMockLogicMessage(apiMsg, clientData)

		GRPCChans.AtDockChannel <- atDockMockLogicMessage

	case a.APIMessageTypeOnBoat:
		Logger.Info().Msgf("Processing %s message to MockLogic from ConnName: %s, ConnType: %s",
			mlMsg.MessageType, mlMsg.ConnName, mlMsg.ConnType)
		clientData := a.NewClientData(mlMsg.ConnName, mlMsg.ConnType)
		var apiMsg a.OnBoatAPIMessage
		err := json.Unmarshal(mlMsg.APIMessageBytes, &apiMsg)
		if err != nil {
			Logger.Debug().Msgf("Error unmarshaling %s message: %s",
				mlMsg.MessageType, err.Error())
			return
		}

		onBoatMockLogicMessage := a.NewOnBoatMockLogicMessage(apiMsg, clientData)

		GRPCChans.OnBoatChannel <- onBoatMockLogicMessage

	case a.APIMessageTypeOffBoat:
		Logger.Info().Msgf("Processing %s message to MockLogic from ConnName: %s, ConnType: %s",
			mlMsg.MessageType, mlMsg.ConnName, mlMsg.ConnType)
		clientData := a.NewClientData(mlMsg.ConnName, mlMsg.ConnType)
		var apiMsg a.OffBoatAPIMessage
		err := json.Unmarshal(mlMsg.APIMessageBytes, &apiMsg)
		if err != nil {
			Logger.Debug().Msgf("Error unmarshaling %s message: %s",
				mlMsg.MessageType, err.Error())
			return
		}

		offBoatMockLogicMessage := a.NewOffBoatMockLogicMessage(apiMsg, clientData)

		GRPCChans.OffBoatChannel <- offBoatMockLogicMessage

	default:
		Logger.Warn().Msgf("Received unknown message type, %s, from ConnName: %s, ConnType: %s",
			mlMsg.MessageType, mlMsg.ConnName, mlMsg.ConnType)

	}
}

// Reserve handles sending and receiving the bi-directional stream for ReserveMessage
func (s *adapterServer) ReserveTrip(stream pb.Adapter_ReserveTripServer) error {
	var err error

	// Launch a goroutine to receive the stream of Empty messages
	// These are not expected to be received from the gRPC client and
	// can be discarded.
	go func() {
		for {
			_, err := stream.Recv()
			if err == io.EOF {
				// read done.
				return
			}
			if err != nil {
				Logger.Debug().Msgf("Failed to receive an Empty message: %s", err.Error())
			}
		}
	}()

	// Wait for messages to appear on the channel and send to the MockLogic
	for reserveMockLogicMsg := range GRPCChans.ReserveTripChannel {
		// Create gRPC ReserveMessage
		apiMessage := pb.ReserveTripAPIMessage{
			MessageType: reserveMockLogicMsg.APIMessage.MessageType,
			AuthToken:   reserveMockLogicMsg.APIMessage.AuthToken,
			ClientID:    reserveMockLogicMsg.APIMessage.ClientID,
			SourceDock: &pb.Dock{
				Address: &pb.Address{
					Number: reserveMockLogicMsg.APIMessage.SourceDock.Address.Number,
					Street: reserveMockLogicMsg.APIMessage.SourceDock.Address.Street,
				},
				Gangway: reserveMockLogicMsg.APIMessage.SourceDock.Gangway,
			},
			DestinationDock: &pb.Dock{
				Address: &pb.Address{
					Number: reserveMockLogicMsg.APIMessage.DestinationDock.Address.Number,
					Street: reserveMockLogicMsg.APIMessage.DestinationDock.Address.Street,
				},
				Gangway: reserveMockLogicMsg.APIMessage.DestinationDock.Gangway,
			},
		}
		clientData := pb.ClientData{
			ConnName: reserveMockLogicMsg.Client.ConnName,
			ConnType: reserveMockLogicMsg.Client.ConnType,
		}
		reserveTripMsg := pb.ReserveTripMessage{
			APIMessage: &apiMessage,
			ClientData: &clientData,
		}

		// Send message to MockLogic
		if err = stream.Send(&reserveTripMsg); err != nil {
			Logger.Debug().Msgf("Failed to send a ReserveMessage: %s", err.Error())
			return err
		}
	}

	return err
}

// AtDock handles sending and receiving the bi-directional stream for AtDockMessage
func (s *adapterServer) AtDock(stream pb.Adapter_AtDockServer) error {
	var err error

	// Launch a goroutine to receive the stream of Empty messages
	// These are not expected to be received from the gRPC client and
	// can be discarded.
	go func() {
		for {
			_, err := stream.Recv()
			if err == io.EOF {
				// read done.
				return
			}
			if err != nil {
				Logger.Debug().Msgf("Failed to receive an Empty message : %v", err)
			}
		}
	}()

	for atDockMockLogicMessage := range GRPCChans.AtDockChannel {
		// Create gRPC AtDockMessage
		apiMessage := pb.AtDockAPIMessage{
			MessageType: atDockMockLogicMessage.APIMessage.MessageType,
			ClientID:    atDockMockLogicMessage.APIMessage.ClientID,
			Boat: &pb.Boat{
				BoatID: atDockMockLogicMessage.APIMessage.Boat.BoatID,
				Name:   atDockMockLogicMessage.APIMessage.Boat.Name,
			},
			Dock: &pb.Dock{
				Address: &pb.Address{
					Number: atDockMockLogicMessage.APIMessage.Dock.Address.Number,
					Street: atDockMockLogicMessage.APIMessage.Dock.Address.Street,
				},
				Gangway: atDockMockLogicMessage.APIMessage.Dock.Gangway,
			},
			TransactionID: atDockMockLogicMessage.APIMessage.TransactionID,
		}
		clientData := pb.ClientData{
			ConnName: atDockMockLogicMessage.Client.ConnName,
			ConnType: atDockMockLogicMessage.Client.ConnType,
		}
		atDockMessage := pb.AtDockMessage{
			APIMessage: &apiMessage,
			ClientData: &clientData,
		}

		// Send message to MockLogic
		if err = stream.Send(&atDockMessage); err != nil {
			Logger.Debug().Msgf("Failed to send an AtDockMessage: %v", err)
			return err
		}
	}

	return err
}

// OnBoat handles sending and receiving the bi-directional stream for OnBoatMessage
func (s *adapterServer) OnBoat(stream pb.Adapter_OnBoatServer) error {
	var err error

	// Launch a goroutine to receive the stream of Empty messages
	// These are not expected to be received from the gRPC client and
	// can be discarded
	go func() {
		for {
			_, err := stream.Recv()
			if err == io.EOF {
				// read done.
				return
			}
			if err != nil {
				Logger.Debug().Msgf("Failed to receive an Empty message : %v", err)
			}
		}
	}()

	for onBoatMockLogicMessage := range GRPCChans.OnBoatChannel {
		// Create gRPC OnBoatMessage
		apiMessage := pb.OnBoatAPIMessage{
			MessageType: onBoatMockLogicMessage.APIMessage.MessageType,
			ClientID:    onBoatMockLogicMessage.APIMessage.ClientID,
			Boat: &pb.Boat{
				BoatID: onBoatMockLogicMessage.APIMessage.Boat.BoatID,
				Name:   onBoatMockLogicMessage.APIMessage.Boat.Name,
			},
			TransactionID: onBoatMockLogicMessage.APIMessage.TransactionID,
		}
		clientData := pb.ClientData{
			ConnName: onBoatMockLogicMessage.Client.ConnName,
			ConnType: onBoatMockLogicMessage.Client.ConnType,
		}
		onBoatMessage := pb.OnBoatMessage{
			APIMessage: &apiMessage,
			ClientData: &clientData,
		}

		// Send message to MockLogic
		if err = stream.Send(&onBoatMessage); err != nil {
			Logger.Debug().Msgf("Failed to send an OnBoatMessage: %v", err)
			return err
		}
	}

	return err
}

// OffBoat handles sending and receiving the bi-directional stream for OffBoatMessage
func (s *adapterServer) OffBoat(stream pb.Adapter_OffBoatServer) error {
	var err error

	// Launch a goroutine to receive the stream of Empty messages
	// These are not expected to be received from the gRPC client and
	// can be discarded
	go func() {
		for {
			_, err := stream.Recv()
			if err == io.EOF {
				// read done.
				return
			}
			if err != nil {
				Logger.Debug().Msgf("Failed to receive an Empty message : %v", err)
			}
		}
	}()

	for offBoatMockLogicMessage := range GRPCChans.OffBoatChannel {
		// Create gRPC OffBoatMessage
		apiMessage := pb.OffBoatAPIMessage{
			MessageType: offBoatMockLogicMessage.APIMessage.MessageType,
			ClientID:    offBoatMockLogicMessage.APIMessage.ClientID,
			Boat: &pb.Boat{
				BoatID: offBoatMockLogicMessage.APIMessage.Boat.BoatID,
				Name:   offBoatMockLogicMessage.APIMessage.Boat.Name,
			},
			TransactionID: offBoatMockLogicMessage.APIMessage.TransactionID,
		}
		clientData := pb.ClientData{
			ConnName: offBoatMockLogicMessage.Client.ConnName,
			ConnType: offBoatMockLogicMessage.Client.ConnType,
		}
		offBoatMessage := pb.OffBoatMessage{
			APIMessage: &apiMessage,
			ClientData: &clientData,
		}

		// Send message to MockLogic
		if err = stream.Send(&offBoatMessage); err != nil {
			Logger.Debug().Msgf("Failed to send an OffBoatMessage: %v", err)
			return err
		}
	}

	return err
}

// Ack handles sending and receiving the bi-directional stream for AckMessage
func (s *adapterServer) Ack(stream pb.Adapter_AckServer) error {
	// No goroutine is launched to write Empty messages since these are not expected
	// by the MockLogic

	// Receive the stream of Ack messages
	for {
		in, err := stream.Recv()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return err
		}

		// Convert pb.AckMessage to MockLogicMessage message
		boat := a.NewBoat(in.APIMessage.Boat.BoatID, in.APIMessage.Boat.Name)
		ackAPIMsg := a.NewAckAPIMessage(a.APIMessageTypeAck, in.APIMessage.ClientID,
			in.APIMessage.IsReserved, boat, in.APIMessage.TransactionID)
		apiMsgBytes, err := json.Marshal(ackAPIMsg)
		if err != nil {
			Logger.Debug().Msgf("Error marshaling %s message received from MockLogic: %s",
				in.APIMessage.MessageType, err.Error())
			continue
		}
		mlMsg := NewMockLogicMessage(in.ClientData.ConnName, in.ClientData.ConnType,
			a.APIMessageTypeAck, apiMsgBytes)

		go ProcessMessageFromMockLogic(&mlMsg)
	}
}

// BoatStatus handles sending and receiving the bi-directional stream for BoatStatusMessage
func (s *adapterServer) BoatStatus(stream pb.Adapter_BoatStatusServer) error {
	// No goroutine is launched to write Empty messages since these are not expected
	// by the MockLogic

	// Receive the stream of BoatStatus messages
	for {
		in, err := stream.Recv()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return err
		}

		// Convert pb.BoatStatusMessage to MockLogicMessage message
		boat := a.NewBoat(in.APIMessage.Boat.BoatID, in.APIMessage.Boat.Name)
		previousDockAddress := a.NewAddress(in.APIMessage.PreviousDock.Address.Number,
			in.APIMessage.PreviousDock.Address.Street)
		previousDock := a.NewDock(previousDockAddress, in.APIMessage.PreviousDock.Gangway)
		currentDockAddress := a.NewAddress(in.APIMessage.CurrentDock.Address.Number,
			in.APIMessage.CurrentDock.Address.Street)
		currentDock := a.NewDock(currentDockAddress, in.APIMessage.CurrentDock.Gangway)
		nextDockAddress := a.NewAddress(in.APIMessage.NextDock.Address.Number,
			in.APIMessage.NextDock.Address.Street)
		nextDock := a.NewDock(nextDockAddress, in.APIMessage.NextDock.Gangway)
		boatStatusAPIMsg := a.NewBoatStatusAPIMessage(a.APIMessageTypeBoatStatus, boat,
			int32(in.APIMessage.ServiceState), previousDock, currentDock, nextDock)
		apiMsgBytes, err := json.Marshal(boatStatusAPIMsg)
		if err != nil {
			Logger.Debug().Msgf("Error marshaling %s message received from MockLogic: %s",
				in.APIMessage.MessageType, err.Error())
			continue
		}
		mlMsg := NewMockLogicMessage(in.ClientData.ConnName, a.ConnectionTypeAll,
			a.APIMessageTypeBoatStatus, apiMsgBytes)

		go ProcessMessageFromMockLogic(&mlMsg)
	}
}

// Arrived handles sending and receiving the bi-directional stream for ArrivedMessage
func (s *adapterServer) Arrived(stream pb.Adapter_ArrivedServer) error {
	// No goroutine is launched to write Empty messages since these are not expected
	// by the MockLogic

	// Receive the stream of Arrived messages
	for {
		in, err := stream.Recv()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return err
		}

		// Convert pb.ArrivedMessage to MockLogicMessage message
		boat := a.NewBoat(in.APIMessage.Boat.BoatID, in.APIMessage.Boat.Name)
		address := a.NewAddress(in.APIMessage.Dock.Address.Number, in.APIMessage.Dock.Address.Street)
		dock := a.NewDock(address, in.APIMessage.Dock.Gangway)
		arrivedAPIMsg := a.NewArrivedAPIMessage(a.APIMessageTypeArrived, in.APIMessage.ClientID,
			boat, dock, in.APIMessage.TransactionID)
		apiMsgBytes, err := json.Marshal(arrivedAPIMsg)
		if err != nil {
			Logger.Debug().Msgf("Error marshaling %s message received from MockLogic: %s",
				in.APIMessage.MessageType, err.Error())
			continue
		}
		mlMsg := NewMockLogicMessage(in.ClientData.ConnName, in.ClientData.ConnType,
			a.APIMessageTypeArrived, apiMsgBytes)

		go ProcessMessageFromMockLogic(&mlMsg)
	}
}

// ProcessMessageToMockLogic processes a message that is being sent from
// the MockLogic to the API clients
func ProcessMessageFromMockLogic(mlMsg *MockLogicMessage) {
	switch mlMsg.ConnType {
	case a.ConnectionTypeWebSocket:
		// Convert message to wss.AdapterMessage
		wssAdapterMsg := wss.NewAdapterMessage(mlMsg.ConnName, mlMsg.APIMessageBytes)
		WebSocketServerConn.Write <- wssAdapterMsg

	case a.ConnectionTypeAll:
		// Convert message to wss.AdapterMessage
		wssAdapterMsg := wss.NewAdapterMessage(wss.WSSServerAllClientsConnName,
			mlMsg.APIMessageBytes)
		WebSocketServerConn.Write <- wssAdapterMsg

		// Convert message to other protocol types here once they are implemented

	default:
		Logger.Warn().Msgf("ProcessMessageFromMockLogic received a message with an unexpected ConnType: %s", mlMsg.ConnType)
	}
}
