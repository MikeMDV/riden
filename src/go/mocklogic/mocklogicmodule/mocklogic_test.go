package main

import (
	"fmt"
	"os"
	a "riden/adapter"
	"riden/logger"
	"testing"

	"github.com/rs/zerolog"
)

// Common test parameters

var simDock1 a.Dock
var simDock2 a.Dock
var simDock3 a.Dock
var simDock4 a.Dock
var simBoat1 a.Boat
var simBoat2 a.Boat

func TestMain(m *testing.M) {
	SetUp()
	retCode := m.Run()
	TearDown()
	os.Exit(retCode)
}

func SetUp() {
	// Initialize logs
	logFile := "./mocklogic_unit_test.log"
	var err error
	Logger, err = logger.InitializeLogger(logFile, zerolog.LevelDebugValue)
	if err != nil {
		fmt.Println("Error opening log file:", logFile, ":", err)
		os.Exit(1)
	}

	// Insert any global data necessary for testing here
	simBoat1 = a.Boat{
		BoatID: 1,
		Name:   "Argo",
	}

	simBoat2 = a.Boat{
		BoatID: 2,
		Name:   "Riverview",
	}

	simDock1 = a.Dock{
		Address: a.Address{
			Number: 901,
			Street: "Carson St",
		},
		Gangway: a.GangwayLocationFore,
	}

	simDock2 = a.Dock{
		Address: a.Address{
			Number: 996,
			Street: "Ohio Rver Blvd",
		},
		Gangway: a.GangwayLocationFore,
	}

	simDock3 = a.Dock{
		Address: a.Address{
			Number: 993,
			Street: "16th St",
		},
		Gangway: a.GangwayLocationFore,
	}

	simDock4 = a.Dock{
		Address: a.Address{
			Number: 964,
			Street: "Stanwix St",
		},
		Gangway: a.GangwayLocationFore,
	}

	SimDockAdjacencyList = BuildDockAdjacencyList()
}

func TearDown() {
	// Delete any global data necessary for testing here

}

func TestReturnDistance(t *testing.T) {
	type testCase struct {
		name             string
		startDock        a.Dock
		goalDock         a.Dock
		expectedDistance int8
	}

	cases := []testCase{
		{
			name:             "ReturnDistance - Same dock for start and goal",
			startDock:        simDock1,
			goalDock:         simDock1,
			expectedDistance: 0,
		},
		{
			name:             "ReturnDistance - Dock 1 to Dock 4",
			startDock:        simDock1,
			goalDock:         simDock4,
			expectedDistance: 3,
		},
		{
			name:             "ReturnDistance - Dock 2 to Dock 3",
			startDock:        simDock2,
			goalDock:         simDock3,
			expectedDistance: 1,
		},
		{
			name:             "ReturnDistance - Dock 2 to Dock 1 - wraparound",
			startDock:        simDock2,
			goalDock:         simDock1,
			expectedDistance: 3,
		},
	}

	for _, testCase := range cases {
		recdDistance := ReturnDistance(testCase.startDock, testCase.goalDock)

		if recdDistance != testCase.expectedDistance {
			t.Fatalf("Expected distance %d but received %d in test case: %s",
				testCase.expectedDistance, recdDistance, testCase.name)
		}
	}
}

func TestAddBoatsInService(t *testing.T) {
	type testCase struct {
		name             string
		boatStatus1      a.BoatStatusAPIMessage
		boatStatus2      a.BoatStatusAPIMessage
		expectedBoatsLen int
	}

	cases := []testCase{
		{
			name: "AddBoatsInService - Two boats with unknown state",
			boatStatus1: a.BoatStatusAPIMessage{
				Boat:         simBoat1,
				ServiceState: a.ServiceStateUnknown,
			},
			boatStatus2: a.BoatStatusAPIMessage{
				Boat:         simBoat2,
				ServiceState: a.ServiceStateUnknown,
			},
			expectedBoatsLen: 0,
		},
		{
			name: "AddBoatsInService - Two boats with on-time state",
			boatStatus1: a.BoatStatusAPIMessage{
				Boat:         simBoat1,
				ServiceState: a.ServiceStateOnTime,
			},
			boatStatus2: a.BoatStatusAPIMessage{
				Boat:         simBoat2,
				ServiceState: a.ServiceStateOnTime,
			},
			expectedBoatsLen: 2,
		},
		{
			name: "AddBoatsInService - Two boats with delayed state",
			boatStatus1: a.BoatStatusAPIMessage{
				Boat:         simBoat1,
				ServiceState: a.ServiceStateDelayed,
			},
			boatStatus2: a.BoatStatusAPIMessage{
				Boat:         simBoat2,
				ServiceState: a.ServiceStateDelayed,
			},
			expectedBoatsLen: 2,
		},
		{
			name: "AddBoatsInService - One boat with delayed state",
			boatStatus1: a.BoatStatusAPIMessage{
				Boat:         simBoat1,
				ServiceState: a.ServiceStateUnavailable,
			},
			boatStatus2: a.BoatStatusAPIMessage{
				Boat:         simBoat2,
				ServiceState: a.ServiceStateDelayed,
			},
			expectedBoatsLen: 1,
		},
		{
			name: "AddBoatsInService - Two boats with unavailable state",
			boatStatus1: a.BoatStatusAPIMessage{
				Boat:         simBoat1,
				ServiceState: a.ServiceStateUnavailable,
			},
			boatStatus2: a.BoatStatusAPIMessage{
				Boat:         simBoat2,
				ServiceState: a.ServiceStateUnavailable,
			},
			expectedBoatsLen: 0,
		},
	}

	for _, testCase := range cases {
		var boats []a.Boat
		safeBoatStatuses.Store(testCase.boatStatus1.Boat.BoatID,
			testCase.boatStatus1)
		safeBoatStatuses.Store(testCase.boatStatus2.Boat.BoatID,
			testCase.boatStatus2)
		AddBoatsInService(&boats)

		if len(boats) != testCase.expectedBoatsLen {
			t.Fatalf("Expected len(boats) %d but received %d in test case: %s",
				testCase.expectedBoatsLen, len(boats), testCase.name)
		}
	}
}

func TestAdvanceToTripState(t *testing.T) {
	type testCase struct {
		name          string
		trip          Trip
		state         int32
		expectedError bool
	}

	cases := []testCase{
		{
			name: "AdvanceToTripState - Normal state transition",
			trip: Trip{
				TripState: TripStateUnknown,
			},
			state:         TripStateReserved,
			expectedError: false,
		},
		{
			name: "AdvanceToTripState - State more than 1 state from current",
			trip: Trip{
				TripState: TripStateUnknown,
			},
			state:         TripStateClientAtDock,
			expectedError: true,
		},
		{
			name: "AdvanceToTripState - State less than current",
			trip: Trip{
				TripState: TripStateClientOnBoat,
			},
			state:         TripStateClientAtDock,
			expectedError: true,
		},
		{
			name: "AdvanceToTripState - State below range",
			trip: Trip{
				TripState: TripStateClientOnBoat,
			},
			state:         -1,
			expectedError: true,
		},
		{
			name: "AdvanceToTripState - State above range",
			trip: Trip{
				TripState: TripStateClientOnBoat,
			},
			state:         10,
			expectedError: true,
		},
	}

	for _, testCase := range cases {
		err := testCase.trip.AdvanceToTripState(testCase.state)

		if testCase.expectedError == (err == nil) {
			t.Fatalf("Expectation of error was %v for test %s but received error was: %v",
				testCase.expectedError, testCase.name, err)
		}
	}
}
