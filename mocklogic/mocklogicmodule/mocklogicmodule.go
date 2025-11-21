package main

import (
	"container/ring"
	"flag"
	"fmt"
	"os"
	"path"
	a "riden/adapter"
	"riden/logger"

	"github.com/rs/zerolog"
)

// Logger Handles all log writing for the MockLogic
var Logger zerolog.Logger

var LogDirectory string
var ParameterDirectory string

var SimFrameRing *ring.Ring

var SimFrameChannel chan []a.BoatStatusAPIMessage

var StopSimFrmaes chan int

var SimBoatStatusChannel chan a.BoatStatusAPIMessage

// InitiateSimFrames builds the sim frame ring and launches the goroutine
// that advances the frames
func InitiateSimFrames() {
	SimFrameRing = BuildSimFrameRing(SimFrames)

	SimFrameChannel = make(chan []a.BoatStatusAPIMessage)
	SimBoatStatusChannel = make(chan a.BoatStatusAPIMessage, SimBoatTotal)

	go AdvanceSimFrames()
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

	logFile := path.Join(LogDirectory, "adapter.log")

	// Startup procedures
	var err error
	Logger, err = logger.InitializeLogger(logFile, flag.Arg(1))
	if err != nil {
		fmt.Println("Error opening log file:", logFile, ":", err.Error())
		os.Exit(1)
	}

	InitiateSimFrames()
}
