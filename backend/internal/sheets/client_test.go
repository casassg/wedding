package sheets

import (
	"testing"

	"github.com/stretchr/testify/require"
)

// TestTravelRoundTrip builds the human-readable labels WriteTravel produces
// for each travel enum combination and feeds them through the parse helpers,
// asserting the original enums come back. This guards against label drift
// between WriteTravel and the ReadSheet parsing added to restore travel data
// on fresh boot.
func TestTravelRoundTrip(t *testing.T) {
	cases := []struct {
		name      string
		busTo     string
		pickup    string
		busReturn string
		hotel     string
		cocktail  string
		brunch    string
	}{
		{name: "thursday sap", busTo: "thursday", pickup: "sap", busReturn: "sunday_san_pedro", hotel: "marina_copan", cocktail: "yes", brunch: "no"},
		{name: "thursday welchez", busTo: "thursday", pickup: "welchez", busReturn: "sunday_sap", hotel: "plaza_copan", cocktail: "no", brunch: "yes"},
		{name: "thursday no pickup", busTo: "thursday", pickup: "", busReturn: "monday_san_pedro", hotel: "plaza_magdalena"},
		{name: "friday sap", busTo: "friday", pickup: "sap", busReturn: "monday_sap", hotel: "yatbalam", cocktail: "yes", brunch: "yes"},
		{name: "friday welchez", busTo: "friday", pickup: "welchez", busReturn: "none"},
		{name: "friday no pickup", busTo: "friday", pickup: ""},
		{name: "no bus", busTo: "none", pickup: ""},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			// Build the labels the same way WriteTravel does.
			busToLabel := ""
			switch tc.busTo {
			case "thursday":
				switch tc.pickup {
				case "sap":
					busToLabel = "Thursday from SAP airport"
				case "welchez":
					busToLabel = "Thursday from Welchez Café"
				default:
					busToLabel = "Thursday"
				}
			case "friday":
				switch tc.pickup {
				case "sap":
					busToLabel = "Friday from SAP airport"
				case "welchez":
					busToLabel = "Friday from Welchez Café"
				default:
					busToLabel = "Friday"
				}
			case "none":
				busToLabel = "No bus"
			}

			busReturnLabel := ""
			switch tc.busReturn {
			case "sunday_san_pedro":
				busReturnLabel = "Sunday → San Pedro"
			case "sunday_sap":
				busReturnLabel = "Sunday → SAP"
			case "monday_san_pedro":
				busReturnLabel = "Monday → San Pedro"
			case "monday_sap":
				busReturnLabel = "Monday → SAP"
			case "none":
				busReturnLabel = "No bus"
			}

			hotelLabel := tc.hotel
			switch tc.hotel {
			case "marina_copan":
				hotelLabel = "Hotel Marina Copan"
			case "plaza_copan":
				hotelLabel = "Hotel Plaza Copan"
			case "plaza_magdalena":
				hotelLabel = "Hotel Plaza Magdalena"
			case "yatbalam":
				hotelLabel = "Hotel Yat B'alam"
			}

			cocktailLabel := ""
			if tc.cocktail == "yes" {
				cocktailLabel = "Yes"
			} else if tc.cocktail == "no" {
				cocktailLabel = "No"
			}
			brunchLabel := ""
			if tc.brunch == "yes" {
				brunchLabel = "Yes"
			} else if tc.brunch == "no" {
				brunchLabel = "No"
			}

			// Feed the labels back through the parse helpers.
			gotBusTo, gotPickup := parseBusTo(busToLabel)
			require.Equal(t, tc.busTo, gotBusTo)
			require.Equal(t, tc.pickup, gotPickup)

			require.Equal(t, tc.busReturn, parseBusReturn(busReturnLabel))
			require.Equal(t, tc.hotel, parseHotel(hotelLabel))
			require.Equal(t, tc.cocktail, parseYesNo(cocktailLabel))
			require.Equal(t, tc.brunch, parseYesNo(brunchLabel))
		})
	}
}

// TestTravelParseFallbacks checks that hand-edited or unknown sheet text
// degrades safely instead of erroring or destroying data.
func TestTravelParseFallbacks(t *testing.T) {
	busTo, pickup := parseBusTo("some hand-typed nonsense")
	require.Equal(t, "", busTo)
	require.Equal(t, "", pickup)

	require.Equal(t, "", parseBusReturn("whatever"))
	require.Equal(t, "", parseYesNo("maybe"))

	// Hand-typed hotels pass through unchanged rather than being wiped.
	require.Equal(t, "Casa de la Tía Rosa", parseHotel("Casa de la Tía Rosa"))
}
