package main

import (
	"container/list"
	"container/ring"
	"context"
	"flag"
	"fmt"
	"io"
	"os"
	"path"
	a "riden/adapter"
	"riden/logger"
	pb "riden/proto"
	"sync"
	"time"

	"github.com/rs/zerolog"
	"google.golang.org/grpc"
)

// Logger Handles all log writing for the MockLogic
var Logger zerolog.Logger

var LogDirectory string

// Simulation frame related
var SimFrameRing *ring.Ring
var StopSimFrames chan int
var SimFrameBoatStatusChannel chan a.BoatStatusAPIMessage
var SimDockAdjacencyList []*list.List

// gRPC related
var GRPCDialTimer *time.Timer
var GRPCStreamWaitChannel chan struct{}
var Once sync.Once
var CloseWaitChan func()
var AdapterAckChannel chan a.AckMockLogicMessage
var AdapterBoatStatusChannel chan a.BoatStatusMockLogicMessage
var AdapterArrivedChannel chan a.ArrivedMockLogicMessage

// InitializeSimFrames builds the sim frame ring and launches the goroutine
// that advances the frames
func InitializeSimFrames() {
	SimFrameRing = BuildSimFrameRing(SimFrames)

	SimFrameBoatStatusChannel = make(chan a.BoatStatusAPIMessage, 2)

	SimDockAdjacencyList = BuildDockAdjacencyList()

	// TODO: Make safeTrips sync.Map [string]Trip to hold reserved trips and
	// create loop to check SimFrameBoatStatusChannel and determine if Arrived
	// should be sent every time a new status is updated.

	go AdvanceSimFrames()
}

func InitializeAdapterGRPCStreams() {
	var opts []grpc.DialOption // No options, currently
	conn, err := grpc.NewClient(a.GRPCServerAddress, opts...)
	if err != nil {
		Logger.Error().Msgf("Failed to dial gRPC: %s", err.Error())
		GRPCDialTimer = time.AfterFunc(time.Duration(2*time.Second), InitializeAdapterGRPCStreams)
		return
	}

	if GRPCDialTimer != nil {
		GRPCDialTimer.Stop()
	}

	// Make channels for incoming messages
	AdapterAckChannel = make(chan a.AckMockLogicMessage)
	AdapterBoatStatusChannel = make(chan a.BoatStatusMockLogicMessage, 2)
	AdapterArrivedChannel = make(chan a.ArrivedMockLogicMessage)

	client := pb.NewAdapterClient(conn)
	ctx, cancel := context.WithCancel(context.Background())
	// Make the wait channel and close function
	GRPCStreamWaitChannel = make(chan struct{})
	CloseWaitChan = func() {
		close(GRPCStreamWaitChannel)
	}

	// Launch streams
	go runReserveTrip(ctx, client)
	go runAck(ctx, client)
	go runAtDock(ctx, client)
	go runOnBoat(ctx, client)
	go runOffBoat(ctx, client)
	go runBoatStatus(ctx, client)
	go runArrived(ctx, client)

	// Block until signaled
	<-GRPCStreamWaitChannel
	// One of the streams has failed, so cancel context, close connection
	// and attempt to reconnect
	cancel()
	conn.Close()
	time.AfterFunc(time.Duration(500*time.Millisecond), InitializeAdapterGRPCStreams)

}

// runReserveTrip handles the ReserveTrip bidi stream. The stream is receiving the ReserveTrip
// messages from the Adapter and is not expected to send any Empty messages to the Adapter,
// so Send() will not be called.
func runReserveTrip(ctx context.Context, client pb.AdapterClient) {
	Logger.Info().Msg("Starting ReserveTrip stream")

	stream, err := client.ReserveTrip(ctx)
	if err != nil {
		Logger.Error().Msgf("client.Reserve failed to create stream: %s", err.Error())
		Once.Do(CloseWaitChan)
	}

	for {
		in, err := stream.Recv()
		if err == io.EOF {
			// Read done
			Logger.Warn().Msg("client.ReserveTrip ended with EOF")
			break
		}
		if err != nil {
			Logger.Error().Msgf("client.ReserveTrip failed: %s", err.Error())
			break
		}
		Logger.Info().Msgf("Received ReserveTripMessage: %q", in)
		// TODO: Handle message
	}

	Once.Do(CloseWaitChan)
	stream.CloseSend()
}

// runAtDock handles the AtDock bidi stream. The stream is receiving the AtDock
// messages from the Adapter and is not expected to send any Empty messages to
// the Adapter, so Send() will not be called.
func runAtDock(ctx context.Context, client pb.AdapterClient) {
	Logger.Info().Msg("Starting AtDock stream")

	stream, err := client.AtDock(ctx)
	if err != nil {
		Logger.Error().Msgf("client.AtDock failed to create stream: %s", err.Error())
		Once.Do(CloseWaitChan)
	}

	for {
		in, err := stream.Recv()
		if err == io.EOF {
			// Read done
			Logger.Warn().Msg("client.AtDock ended with EOF")
			break
		}
		if err != nil {
			Logger.Error().Msgf("client.AtDock failed: %s", err.Error())
			break
		}
		Logger.Info().Msgf("Received AtDockMessage: %q", in)
		// TODO: Handle message
	}

	Once.Do(CloseWaitChan)
	stream.CloseSend()
}

// runOnBoat handles the OnBoat bidi stream. The stream is receiving the OnBoat
// messages from the Adapter and is not expected to send any Empty messages to
// the Adapter, so Send() will not be called.
func runOnBoat(ctx context.Context, client pb.AdapterClient) {
	Logger.Info().Msg("Starting OnBoat stream")

	stream, err := client.OnBoat(ctx)
	if err != nil {
		Logger.Error().Msgf("client.OnBoat failed to create stream: %s", err.Error())
		Once.Do(CloseWaitChan)
	}

	for {
		in, err := stream.Recv()
		if err == io.EOF {
			// Read done
			Logger.Warn().Msg("client.OnBoat ended with EOF")
			break
		}
		if err != nil {
			Logger.Error().Msgf("client.OnBoat failed: %s", err.Error())
			break
		}
		Logger.Info().Msgf("Received OnBoatMessage: %q", in)
		// TODO: Handle message
	}

	Once.Do(CloseWaitChan)
	stream.CloseSend()
}

// runOffBoat handles the OffBoat bidi stream. The stream is receiving the OffBoat
// messages from the Adapter and is not expected to send any Empty messages to
// the Adapter, so Send() will not be called.
func runOffBoat(ctx context.Context, client pb.AdapterClient) {
	Logger.Info().Msg("Starting OffBoat stream")

	stream, err := client.OffBoat(ctx)
	if err != nil {
		Logger.Error().Msgf("client.OffBoat failed to create stream: %s", err.Error())
		Once.Do(CloseWaitChan)
	}

	for {
		in, err := stream.Recv()
		if err == io.EOF {
			// Read done
			Logger.Warn().Msg("client.OffBoat ended with EOF")
			break
		}
		if err != nil {
			Logger.Error().Msgf("client.OffBoat failed: %s", err.Error())
			break
		}
		Logger.Info().Msgf("Received OffBoatMessage: %q", in)
		// TODO: Handle message
	}

	Once.Do(CloseWaitChan)
	stream.CloseSend()
}

// runAck handles the Ack bidi stream. The stream is sending the Ack messages to the
// Adapter and is not expected to receive any Empty messages from the Adapter, so
// Recv() will not be called.
func runAck(ctx context.Context, client pb.AdapterClient) {
	Logger.Info().Msg("Starting ReserveTrip stream")

	stream, err := client.Ack(ctx)
	if err != nil {
		Logger.Error().Msgf("client.Ack failed to create stream: %s", err.Error())
		Once.Do(CloseWaitChan)
	}

	select {
	case <-ctx.Done():
		Logger.Warn().Msgf("client.Ack context canceled with err: %s", ctx.Err().Error())
		return

	case ack := <-AdapterAckChannel:
		ackBoatGRPC := pb.Boat{
			BoatId: ack.APIMessage.Boat.BoatID,
			Name:   ack.APIMessage.Boat.Name,
		}
		ackAPIMessageGRPC := pb.AckAPIMessage{
			MessageType:   ack.APIMessage.MessageType,
			ClientId:      ack.APIMessage.ClientID,
			IsReserved:    ack.APIMessage.IsReserved,
			Boat:          &ackBoatGRPC,
			TransactionId: ack.APIMessage.TransactionID,
		}
		ackClientDataGRPC := pb.ClientData{
			ConnName: ack.Client.ConnName,
			ConnType: ack.Client.ConnType,
		}
		ackMessageGRPC := pb.AckMessage{
			ApiMessage: &ackAPIMessageGRPC,
			ClientData: &ackClientDataGRPC,
		}

		if err := stream.Send(&ackMessageGRPC); err != nil {
			Logger.Error().Msgf("client.Ack failed to send msg: %s", err.Error())
			Once.Do(CloseWaitChan)
			return
		}
	}
}

// runBoatStatus handles the BoatStatus bidi stream. The stream is sending the BoatStatus messages to the
// Adapter and is not expected to receive any Empty messages from the Adapter, so
// Recv() will not be called.
func runBoatStatus(ctx context.Context, client pb.AdapterClient) {
	Logger.Info().Msg("Starting ReserveTrip stream")

	stream, err := client.BoatStatus(ctx)
	if err != nil {
		Logger.Error().Msgf("client.BoatStatus failed to create stream: %s", err.Error())
		Once.Do(CloseWaitChan)
	}

	select {
	case <-ctx.Done():
		Logger.Warn().Msgf("client.BoatStatus context canceled with err: %s", ctx.Err().Error())
		return

	case status := <-AdapterBoatStatusChannel:
		statusBoatGRPC := pb.Boat{
			BoatId: status.APIMessage.Boat.BoatID,
			Name:   status.APIMessage.Boat.Name,
		}
		statusPrevDockAddress := pb.Address{
			Number: status.APIMessage.PreviousDock.Address.Number,
			Street: status.APIMessage.PreviousDock.Address.Street,
		}
		statusPrevDock := pb.Dock{
			Address: &statusPrevDockAddress,
			Gangway: status.APIMessage.PreviousDock.Gangway,
		}
		statusCurrDockAddress := pb.Address{
			Number: status.APIMessage.CurrentDock.Address.Number,
			Street: status.APIMessage.CurrentDock.Address.Street,
		}
		statusCurrDock := pb.Dock{
			Address: &statusCurrDockAddress,
			Gangway: status.APIMessage.CurrentDock.Gangway,
		}
		statusNextDockAddress := pb.Address{
			Number: status.APIMessage.NextDock.Address.Number,
			Street: status.APIMessage.NextDock.Address.Street,
		}
		statusNextDock := pb.Dock{
			Address: &statusNextDockAddress,
			Gangway: status.APIMessage.NextDock.Gangway,
		}
		statusAPIMessageGRPC := pb.BoatStatusAPIMessage{
			MessageType:  status.APIMessage.MessageType,
			Boat:         &statusBoatGRPC,
			ServiceState: pb.ServiceState(status.APIMessage.ServiceState),
			PreviousDock: &statusPrevDock,
			CurrentDock:  &statusCurrDock,
			NextDock:     &statusNextDock,
		}
		statusClientDataGRPC := pb.ClientData{
			ConnName: status.Client.ConnName,
			ConnType: status.Client.ConnType,
		}
		statusMessageGRPC := pb.BoatStatusMessage{
			ApiMessage: &statusAPIMessageGRPC,
			ClientData: &statusClientDataGRPC,
		}

		if err := stream.Send(&statusMessageGRPC); err != nil {
			Logger.Error().Msgf("client.BoatStatus failed to send msg: %s", err.Error())
			Once.Do(CloseWaitChan)
			return
		}
	}
}

// runArrived handles the Arrived bidi stream. The stream is sending the Arrived messages to the
// Adapter and is not expected to receive any Empty messages from the Adapter, so
// Recv() will not be called.
func runArrived(ctx context.Context, client pb.AdapterClient) {
	Logger.Info().Msg("Starting ReserveTrip stream")

	stream, err := client.Arrived(ctx)
	if err != nil {
		Logger.Error().Msgf("client.Arrived failed to create stream: %s", err.Error())
		Once.Do(CloseWaitChan)
	}

	select {
	case <-ctx.Done():
		Logger.Warn().Msgf("client.Arrived context canceled with err: %s", ctx.Err().Error())
		return

	case arr := <-AdapterArrivedChannel:
		arrBoatGRPC := pb.Boat{
			BoatId: arr.APIMessage.Boat.BoatID,
			Name:   arr.APIMessage.Boat.Name,
		}
		arrDockAddress := pb.Address{
			Number: arr.APIMessage.Dock.Address.Number,
			Street: arr.APIMessage.Dock.Address.Street,
		}
		arrDock := pb.Dock{
			Address: &arrDockAddress,
			Gangway: arr.APIMessage.Dock.Gangway,
		}
		arrAPIMessageGRPC := pb.ArrivedAPIMessage{
			MessageType:   arr.APIMessage.MessageType,
			ClientId:      arr.APIMessage.ClientID,
			Boat:          &arrBoatGRPC,
			Dock:          &arrDock,
			TransactionId: arr.APIMessage.TransactionID,
		}
		arrClientDataGRPC := pb.ClientData{
			ConnName: arr.Client.ConnName,
			ConnType: arr.Client.ConnType,
		}
		arrMessageGRPC := pb.ArrivedMessage{
			ApiMessage: &arrAPIMessageGRPC,
			ClientData: &arrClientDataGRPC,
		}

		if err := stream.Send(&arrMessageGRPC); err != nil {
			Logger.Error().Msgf("client.Arrived failed to send msg: %s", err.Error())
			Once.Do(CloseWaitChan)
			return
		}
	}
}

func Usage() {
	fmt.Println("Usage:", os.Args[0], "log_dir log_level")
	os.Exit(1) // 1 - Non-zero exit code indicates an error
}

func main() {
	// Parse the arguments
	flag.Parse()

	if len(flag.Args()) != 2 {
		Usage()
	}

	LogDirectory = flag.Arg(0)

	fmt.Printf("Log Directory: %s\n", LogDirectory)

	logFile := path.Join(LogDirectory, "mocklogic.log")

	// Startup procedures
	var err error
	Logger, err = logger.InitializeLogger(logFile, flag.Arg(1))
	if err != nil {
		fmt.Println("Error opening log file:", logFile, ":", err.Error())
		os.Exit(1)
	}

	InitializeSimFrames()

	go InitializeAdapterGRPCStreams()

	// Block
	select {}
}
