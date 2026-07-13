package api

import (
	"strings"
	"testing"
	"time"

	"github.com/casassg/wedding/backend/internal/store"
	"github.com/stretchr/testify/require"
)

func TestValidateTravel_ValidInputs(t *testing.T) {
	cases := []TravelRequest{
		{BusTo: "", Pickup: "", BusReturn: ""},
		{BusTo: "thursday", Pickup: "sap", ArrivalFlight: "AV 620", BusReturn: "sunday_sap"},
		{BusTo: "friday", Pickup: "welchez", BusReturn: "monday_san_pedro"},
		{BusTo: "none", BusReturn: "none"},
		{BusTo: "thursday", Pickup: "sap", BusReturn: "sunday_san_pedro", Hotel: "marina", Notes: "ok"},
		{BusTo: "friday", Pickup: "sap", BusReturn: "monday_sap"},
	}
	for _, req := range cases {
		require.NoError(t, validateTravel(req))
	}
}

func TestValidateTravel_InvalidEnums(t *testing.T) {
	cases := []struct {
		name string
		req  TravelRequest
	}{
		{"bad bus_to", TravelRequest{BusTo: "saturday"}},
		{"bad pickup", TravelRequest{Pickup: "hotel"}},
		{"bad bus_return", TravelRequest{BusReturn: "tuesday_sap"}},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			require.Error(t, validateTravel(c.req))
		})
	}
}

func TestValidateTravel_TextLengthLimits(t *testing.T) {
	long := strings.Repeat("x", 501)
	cases := []struct {
		name string
		req  TravelRequest
	}{
		{"flight too long", TravelRequest{ArrivalFlight: long}},
		{"hotel too long", TravelRequest{Hotel: long}},
		{"notes too long", TravelRequest{Notes: long}},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			require.Error(t, validateTravel(c.req))
		})
	}

	require.NoError(t, validateTravel(TravelRequest{Notes: strings.Repeat("é", 500)}))
	require.Error(t, validateTravel(TravelRequest{Notes: strings.Repeat("é", 501)}))
}

func TestToInviteResponse_InHonduras(t *testing.T) {
	cases := []struct {
		location   string
		inHonduras bool
	}{
		{"HONDURAS", true},
		{"honduras", true},
		{"Honduras", true},
		{"", false},
		{"SAP", false},
		{"SPAIN", false},
	}
	for _, c := range cases {
		inv := &store.Invite{Location: c.location}
		resp := ToInviteResponse(inv)
		require.Equal(t, c.inHonduras, resp.InHonduras)
	}
}

func TestToInviteResponse_HasTravelInfo(t *testing.T) {
	inv := &store.Invite{}
	resp := ToInviteResponse(inv)
	require.False(t, resp.HasTravelInfo)

	now := time.Now()
	inv2 := &store.Invite{TravelUpdatedAt: &now}
	resp2 := ToInviteResponse(inv2)
	require.True(t, resp2.HasTravelInfo)
}
