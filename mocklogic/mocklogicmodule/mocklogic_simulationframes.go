package main

import (
	a "riden/adapter"
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

type SimulationFrames struct {
	BoatLocations [16][2]a.BoatStatusAPIMessage
}

// SimFrames simulates two boats moving between 4 docks. The total
// number of frames is 16, with the boats remaining stationary at the
// docks for two frames each and spending two frames traveling
// between each dock.
var SimFrames = SimulationFrames{
	BoatLocations: [16][2]a.BoatStatusAPIMessage{
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
