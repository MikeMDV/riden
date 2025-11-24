package main

import (
	"fmt"
	"math"
	a "riden/adapter"
	"strconv"
	"sync"
	"time"
)

const (
	TripStateUnknown             int32 = 0
	TripStateReserved            int32 = 1
	TripStateClientAtDock        int32 = 2
	TripStateBoatArrivedAtSource int32 = 3
	TripStateClientOnBoat        int32 = 4
	TripStateBoatArrivedAtDest   int32 = 5
	TripStateClientOffBoat       int32 = 6
)

type Trip struct {
	Reservation   a.ReserveTripAPIMessage
	TransactionID string
	Boat          a.Boat
	ServiceState  int32
	TripState     int32
}

// NewTrip returns a new Trip with the fields populated and returns
// an error if the reservation could not be made
func NewTrip(reservation a.ReserveTripAPIMessage) (*Trip, error) {
	var trip = Trip{
		Reservation: reservation,
	}
	trip.GenerateTransactionID()
	err := trip.GetBoatAndServiceStateForTrip()
	if err != nil {
		return &trip, err
	}
	err = trip.AdvanceToTripState(TripStateReserved)
	if err != nil {
		return &trip, err
	}

	return &trip, err
}

func (t *Trip) GenerateTransactionID() {
	transactionNum := time.Now().UnixMilli()
	t.TransactionID = strconv.Itoa(int(transactionNum))
}

func (t *Trip) GetBoatAndServiceStateForTrip() error {
	var err error
	var potentialBoats []a.Boat
	// Check the service states of current boats and add to potential boats
	// if they are on-time or delayed
	AddBoatsInService(&potentialBoats)
	if len(potentialBoats) == 0 {
		err = fmt.Errorf("no boats were available for service")
		return err
	}
	// Check which boat is closest to the SourceDock and set as the boat for
	// the trip and set the service state of the boat.
	closestBoat, err := ReturnClosestBoat(t.Reservation.SourceDock, &potentialBoats)
	if err != nil {
		return err
	}
	t.Boat = closestBoat
	boatStatusVal, ok := safeBoatStatuses.Load(t.Boat.BoatID)
	if ok {
		boatStatus := boatStatusVal.(a.BoatStatusAPIMessage)
		t.ServiceState = boatStatus.ServiceState
	} else {
		err = fmt.Errorf("could not retrieve boat status for ID: %d", t.Boat.BoatID)
	}

	return err
}

// AdvanceToTripState checks the current trip state and advances it to the given
// state if that is permitted
func (t *Trip) AdvanceToTripState(state int32) error {
	var err error
	if state-1 != t.TripState {
		err = fmt.Errorf("current state: %d, cannot advance to state: %d", t.TripState, state)
		return err
	}

	if state > TripStateClientOffBoat || state < TripStateReserved {
		err = fmt.Errorf("state: %d, is out of range", state)
		return err
	}

	t.TripState = state

	return err
}

// safeBoatStatuses holds the statuses of the boats in the system in a
// [int32]BoatStatusAPIMessage map
var safeBoatStatuses sync.Map

func AddBoatsInService(boats *[]a.Boat) {
	safeBoatStatuses.Range(func(key, statusVal interface{}) bool {
		status := statusVal.(a.BoatStatusAPIMessage)
		if status.ServiceState == a.ServiceStateOnTime ||
			status.ServiceState == a.ServiceStateDelayed {
			*boats = append(*boats, status.Boat)
		}

		return true
	})
}

// ReturnClosestBoat checks which boat is closest to the given dock and
// returns it. The check should use the NextDock field in the
// BoatStatusAPIMessage. If the CurrentDock of the boat is the same as
// the SourceDock, boarding has begun and it is too late to reserve
// this boat.
func ReturnClosestBoat(dock a.Dock, boats *[]a.Boat) (a.Boat, error) {
	var err error
	var shortestDistance int8 = math.MaxInt8
	var closestBoat a.Boat
	var boatStatus a.BoatStatusAPIMessage
	for _, boat := range *boats {
		boatStatusVal, ok := safeBoatStatuses.Load(boat.BoatID)
		if ok {
			boatStatus = boatStatusVal.(a.BoatStatusAPIMessage)
		} else {
			err = fmt.Errorf("could not retrieve boat status for ID: %d", boat.BoatID)
		}
		distance := ReturnDistance(boatStatus.NextDock, dock)
		if distance < shortestDistance {
			closestBoat = boat
		}
	}

	return closestBoat, err

}

// ReturnDistance returns an integer representing the distance betweeen the start
// and goal docks
func ReturnDistance(start, goal a.Dock) int8 {
	if start == goal {
		return 0
	}

	// Get the next dock after a and recursivly call
	// ReturnDistance()
	var nextDock a.Dock
	for _, dockList := range SimDockAdjacencyList {
		if dockList.Front().Value == start {
			nextDock = dockList.Front().Next().Value.(a.Dock)
		}
	}

	return 1 + ReturnDistance(nextDock, goal)
}
