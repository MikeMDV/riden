package main

import (
	"container/ring"
	a "riden/adapter"
	wss "riden/websocketserver"
	"time"
)

var SimBoat1 = a.Boat{
	BoatID: 1,
	Name:   "Argo",
}

var SimBoat2 = a.Boat{
	BoatID: 2,
	Name:   "Riverview",
}

var SimDock1 = a.Dock{
	Address: a.Address{
		Number: 901,
		Street: "Carson St",
	},
	Gangway: a.GangwayLocationFore,
}

var SimDock2 = a.Dock{
	Address: a.Address{
		Number: 996,
		Street: "Ohio Rver Blvd",
	},
	Gangway: a.GangwayLocationFore,
}

var SimDock3 = a.Dock{
	Address: a.Address{
		Number: 993,
		Street: "16th St",
	},
	Gangway: a.GangwayLocationFore,
}

var SimDock4 = a.Dock{
	Address: a.Address{
		Number: 964,
		Street: "Stanwix St",
	},
	Gangway: a.GangwayLocationFore,
}

const SimFrameTotal int = 16

const SimBoatTotal int = 2

const SimFrameDuration time.Duration = 15 * time.Second

type SimulationFrames struct {
	BoatLocations [SimFrameTotal][SimBoatTotal]a.BoatStatusAPIMessage
}

// SimFrames simulates two boats moving between 4 docks. The total
// number of frames is 16, with the boats remaining stationary at the
// docks for two frames each and spending two frames traveling
// between each dock.
var SimFrames = SimulationFrames{
	BoatLocations: [SimFrameTotal][SimBoatTotal]a.BoatStatusAPIMessage{
		{
			{
				MessageType:  a.APIMessageTypeBoatStatus,
				Boat:         SimBoat1,
				ServiceState: a.ServiceStateOnTime,
				PreviousDock: SimDock2,
				CurrentDock:  SimDock3,
				NextDock:     SimDock4,
			},
			{
				MessageType:  a.APIMessageTypeBoatStatus,
				Boat:         SimBoat2,
				ServiceState: a.ServiceStateOnTime,
				PreviousDock: SimDock4,
				CurrentDock:  SimDock1,
				NextDock:     SimDock2,
			},
		},
		{
			{
				MessageType:  a.APIMessageTypeBoatStatus,
				Boat:         SimBoat1,
				ServiceState: a.ServiceStateOnTime,
				PreviousDock: SimDock2,
				CurrentDock:  SimDock3,
				NextDock:     SimDock4,
			},
			{
				MessageType:  a.APIMessageTypeBoatStatus,
				Boat:         SimBoat2,
				ServiceState: a.ServiceStateOnTime,
				PreviousDock: SimDock4,
				CurrentDock:  SimDock1,
				NextDock:     SimDock2,
			},
		},
		{
			{
				MessageType:  a.APIMessageTypeBoatStatus,
				Boat:         SimBoat1,
				ServiceState: a.ServiceStateOnTime,
				PreviousDock: SimDock3,
				NextDock:     SimDock4,
			},
			{
				MessageType:  a.APIMessageTypeBoatStatus,
				Boat:         SimBoat2,
				ServiceState: a.ServiceStateOnTime,
				PreviousDock: SimDock1,
				NextDock:     SimDock2,
			},
		},
		{
			{
				MessageType:  a.APIMessageTypeBoatStatus,
				Boat:         SimBoat1,
				ServiceState: a.ServiceStateOnTime,
				PreviousDock: SimDock3,
				NextDock:     SimDock4,
			},
			{
				MessageType:  a.APIMessageTypeBoatStatus,
				Boat:         SimBoat2,
				ServiceState: a.ServiceStateOnTime,
				PreviousDock: SimDock1,
				NextDock:     SimDock2,
			},
		},
		{
			{
				MessageType:  a.APIMessageTypeBoatStatus,
				Boat:         SimBoat1,
				ServiceState: a.ServiceStateOnTime,
				PreviousDock: SimDock3,
				CurrentDock:  SimDock4,
				NextDock:     SimDock1,
			},
			{
				MessageType:  a.APIMessageTypeBoatStatus,
				Boat:         SimBoat2,
				ServiceState: a.ServiceStateOnTime,
				PreviousDock: SimDock1,
				CurrentDock:  SimDock2,
				NextDock:     SimDock3,
			},
		},
		{
			{
				MessageType:  a.APIMessageTypeBoatStatus,
				Boat:         SimBoat1,
				ServiceState: a.ServiceStateOnTime,
				PreviousDock: SimDock3,
				CurrentDock:  SimDock4,
				NextDock:     SimDock1,
			},
			{
				MessageType:  a.APIMessageTypeBoatStatus,
				Boat:         SimBoat2,
				ServiceState: a.ServiceStateOnTime,
				PreviousDock: SimDock1,
				CurrentDock:  SimDock2,
				NextDock:     SimDock3,
			},
		},
		{
			{
				MessageType:  a.APIMessageTypeBoatStatus,
				Boat:         SimBoat1,
				ServiceState: a.ServiceStateOnTime,
				PreviousDock: SimDock4,
				NextDock:     SimDock1,
			},
			{
				MessageType:  a.APIMessageTypeBoatStatus,
				Boat:         SimBoat2,
				ServiceState: a.ServiceStateOnTime,
				PreviousDock: SimDock2,
				NextDock:     SimDock3,
			},
		},
		{
			{
				MessageType:  a.APIMessageTypeBoatStatus,
				Boat:         SimBoat1,
				ServiceState: a.ServiceStateOnTime,
				PreviousDock: SimDock4,
				NextDock:     SimDock1,
			},
			{
				MessageType:  a.APIMessageTypeBoatStatus,
				Boat:         SimBoat2,
				ServiceState: a.ServiceStateOnTime,
				PreviousDock: SimDock2,
				NextDock:     SimDock3,
			},
		},
		{
			{
				MessageType:  a.APIMessageTypeBoatStatus,
				Boat:         SimBoat1,
				ServiceState: a.ServiceStateOnTime,
				PreviousDock: SimDock4,
				CurrentDock:  SimDock1,
				NextDock:     SimDock2,
			},
			{
				MessageType:  a.APIMessageTypeBoatStatus,
				Boat:         SimBoat2,
				ServiceState: a.ServiceStateOnTime,
				PreviousDock: SimDock2,
				CurrentDock:  SimDock3,
				NextDock:     SimDock4,
			},
		},
		{
			{
				MessageType:  a.APIMessageTypeBoatStatus,
				Boat:         SimBoat1,
				ServiceState: a.ServiceStateOnTime,
				PreviousDock: SimDock4,
				CurrentDock:  SimDock1,
				NextDock:     SimDock2,
			},
			{
				MessageType:  a.APIMessageTypeBoatStatus,
				Boat:         SimBoat2,
				ServiceState: a.ServiceStateOnTime,
				PreviousDock: SimDock2,
				CurrentDock:  SimDock3,
				NextDock:     SimDock4,
			},
		},
		{
			{
				MessageType:  a.APIMessageTypeBoatStatus,
				Boat:         SimBoat1,
				ServiceState: a.ServiceStateOnTime,
				PreviousDock: SimDock1,
				NextDock:     SimDock2,
			},
			{
				MessageType:  a.APIMessageTypeBoatStatus,
				Boat:         SimBoat2,
				ServiceState: a.ServiceStateOnTime,
				PreviousDock: SimDock3,
				NextDock:     SimDock4,
			},
		},
		{
			{
				MessageType:  a.APIMessageTypeBoatStatus,
				Boat:         SimBoat1,
				ServiceState: a.ServiceStateOnTime,
				PreviousDock: SimDock1,
				NextDock:     SimDock2,
			},
			{
				MessageType:  a.APIMessageTypeBoatStatus,
				Boat:         SimBoat2,
				ServiceState: a.ServiceStateOnTime,
				PreviousDock: SimDock3,
				NextDock:     SimDock4,
			},
		},
		{
			{
				MessageType:  a.APIMessageTypeBoatStatus,
				Boat:         SimBoat1,
				ServiceState: a.ServiceStateOnTime,
				PreviousDock: SimDock1,
				CurrentDock:  SimDock2,
				NextDock:     SimDock3,
			},
			{
				MessageType:  a.APIMessageTypeBoatStatus,
				Boat:         SimBoat2,
				ServiceState: a.ServiceStateOnTime,
				PreviousDock: SimDock3,
				CurrentDock:  SimDock4,
				NextDock:     SimDock1,
			},
		},
		{
			{
				MessageType:  a.APIMessageTypeBoatStatus,
				Boat:         SimBoat1,
				ServiceState: a.ServiceStateOnTime,
				PreviousDock: SimDock1,
				CurrentDock:  SimDock2,
				NextDock:     SimDock3,
			},
			{
				MessageType:  a.APIMessageTypeBoatStatus,
				Boat:         SimBoat2,
				ServiceState: a.ServiceStateOnTime,
				PreviousDock: SimDock3,
				CurrentDock:  SimDock4,
				NextDock:     SimDock1,
			},
		},
		{
			{
				MessageType:  a.APIMessageTypeBoatStatus,
				Boat:         SimBoat1,
				ServiceState: a.ServiceStateOnTime,
				PreviousDock: SimDock2,
				NextDock:     SimDock3,
			},
			{
				MessageType:  a.APIMessageTypeBoatStatus,
				Boat:         SimBoat2,
				ServiceState: a.ServiceStateOnTime,
				PreviousDock: SimDock4,
				NextDock:     SimDock1,
			},
		},
		{
			{
				MessageType:  a.APIMessageTypeBoatStatus,
				Boat:         SimBoat1,
				ServiceState: a.ServiceStateOnTime,
				PreviousDock: SimDock2,
				NextDock:     SimDock3,
			},
			{
				MessageType:  a.APIMessageTypeBoatStatus,
				Boat:         SimBoat2,
				ServiceState: a.ServiceStateOnTime,
				PreviousDock: SimDock4,
				NextDock:     SimDock1,
			},
		},
	},
}

// BuildSimFrameRing places the SimulationFrames in a Ring container.
// This circular container allows the frames to be advanced through
// all the frames and wrap around to the beginning by repeatedly
// calling Next()
func BuildSimFrameRing(simFrames SimulationFrames) *ring.Ring {
	frameRing := ring.New(len(simFrames.BoatLocations))

	// Get the length of the ring
	n := frameRing.Len()

	// Initialize the ring with SimulationFrames.BoatLocations
	for i := range n {
		frameRing.Value = simFrames.BoatLocations[i]
		frameRing = frameRing.Next()
	}

	return frameRing
}

// AdvanceSimFrames advances the sim frames and places the current frame
// on the SimBoatStatus channel
func AdvanceSimFrames() {
	Logger.Info().Msg("EnteredAdvanceSimFrames()")
	advance := time.Tick(SimFrameDuration)
	for {
		select {
		case <-advance:
			Logger.Info().Msg("Pushing sim frame to channel and advancing")
			statuses := SimFrameRing.Value.([]a.BoatStatusAPIMessage)
			for _, boatStatusAPI := range statuses {
				boatStatus := a.BoatStatusMockLogicMessage{
					APIMessage: boatStatusAPI,
					Client: a.ClientData{
						ConnName: wss.WSSServerAllClientsConnName,
						ConnType: a.ConnectionTypeAll,
					},
				}
				AdapterBoatStatusChannel <- boatStatus
			}
			SimFrameRing = SimFrameRing.Next()

		case stopSignal := <-StopSimFrmaes:
			Logger.Info().Msgf("AdvanceSimFrames has received a stop signal, %d",
				stopSignal)
			// handle close
			return
		}
	}
}
