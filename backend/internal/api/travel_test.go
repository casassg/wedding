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
		{BusTo: "thursday", Pickup: "sap", ArrivalFlight: "AV 620", BusReturn: "sunday_morning_sap"},
		{BusTo: "friday", Pickup: "welchez", BusReturn: "sunday_afternoon_san_pedro"},
		{BusTo: "none", BusReturn: "none"},
		{BusTo: "thursday", Pickup: "sap", BusReturn: "sunday_afternoon_san_pedro", Hotel: "marina", Notes: "ok"},
		{BusTo: "friday", Pickup: "sap", BusReturn: "sunday_morning_sap"},
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
		{"removed monday_san_pedro bus_return", TravelRequest{BusReturn: "monday_san_pedro"}},
		{"removed monday_sap bus_return", TravelRequest{BusReturn: "monday_sap"}},
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
		{"return_detail too long", TravelRequest{ReturnDetail: long}},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			require.Error(t, validateTravel(c.req))
		})
	}

	require.NoError(t, validateTravel(TravelRequest{Notes: strings.Repeat("é", 500)}))
	require.Error(t, validateTravel(TravelRequest{Notes: strings.Repeat("é", 501)}))
}

func TestNormalizeTravel_ClearsReturnDetailWhenNoReturnBus(t *testing.T) {
	req := TravelRequest{BusReturn: "none", ReturnDetail: "UA1422 · Sun, Dec 20 12:30"}
	got := normalizeTravel(req, false)
	require.Equal(t, "", got.ReturnDetail)

	req2 := TravelRequest{BusReturn: "", ReturnDetail: "some hotel"}
	got2 := normalizeTravel(req2, false)
	require.Equal(t, "", got2.ReturnDetail)
}

func TestNormalizeTravel_KeepsReturnDetailForSundaySapOrSanPedro(t *testing.T) {
	req := TravelRequest{BusReturn: "sunday_morning_sap", ReturnDetail: "UA1422 · Sun, Dec 20 12:30"}
	got := normalizeTravel(req, false)
	require.Equal(t, "UA1422 · Sun, Dec 20 12:30", got.ReturnDetail)

	req2 := TravelRequest{BusReturn: "sunday_afternoon_san_pedro", ReturnDetail: "Hotel Marina Copan"}
	got2 := normalizeTravel(req2, false)
	require.Equal(t, "Hotel Marina Copan", got2.ReturnDetail)
}

func TestNormalizeTravel_ClearsEverythingForHonduras(t *testing.T) {
	req := TravelRequest{
		BusTo: "thursday", Pickup: "sap", ArrivalFlight: "UA1422", BusReturn: "sunday_morning_sap",
		Cocktail: "yes", ReturnDetail: "UA1422 · Sun, Dec 20 12:30",
	}
	got := normalizeTravel(req, true)
	require.Equal(t, TravelRequest{Cocktail: ""}, got)
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
