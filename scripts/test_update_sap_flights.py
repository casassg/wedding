import importlib.util
import unittest
from datetime import date, datetime, timezone
from pathlib import Path
from unittest import mock


SCRIPT = Path(__file__).with_name("update-sap-flights.py")
SPEC = importlib.util.spec_from_file_location("update_sap_flights", SCRIPT)
MODULE = importlib.util.module_from_spec(SPEC)
SPEC.loader.exec_module(MODULE)


class UpdateSapFlightsTest(unittest.TestCase):
    def test_fetch_uses_one_scheduled_arrivals_request(self):
        start = datetime(2026, 12, 17, 6, tzinfo=timezone.utc)
        end = datetime(2026, 12, 19, 6, tzinfo=timezone.utc)
        response = {"links": {"next": None}, "scheduled_arrivals": [{"ident": "UAL123"}]}

        with mock.patch.object(MODULE, "api_get", return_value=response) as api_get:
            flights = MODULE.fetch_scheduled_arrivals("key", start, end)

        self.assertEqual([{"ident": "UAL123"}], flights)
        api_get.assert_called_once_with(
            "key",
            "airports/SAP/flights/scheduled_arrivals",
            {
                "start": "2026-12-17T06:00:00Z",
                "end": "2026-12-19T06:00:00Z",
                "type": "Airline",
                "max_pages": 10,
            },
        )

    def test_fetch_refuses_incomplete_response(self):
        response = {"links": {"next": "/more"}, "scheduled_arrivals": []}
        with mock.patch.object(MODULE, "api_get", return_value=response):
            with self.assertRaisesRegex(RuntimeError, "more than 10 pages"):
                MODULE.fetch_scheduled_arrivals("key", datetime.now(timezone.utc), datetime.now(timezone.utc))

    def test_build_arrivals_uses_inline_metadata_and_combines_days(self):
        flights = [
            {
                "operator_iata": "UX",
                "flight_number": "15",
                "origin": {"city": "Madrid", "code_iata": "MAD"},
                "estimated_on": "2026-12-17T18:00:00Z",
                "cancelled": False,
            },
            {
                "operator_iata": "UX",
                "flight_number": "15",
                "origin": {"city": "Madrid", "code_iata": "MAD"},
                "estimated_on": "2026-12-18T18:00:00Z",
                "cancelled": False,
            },
            {
                "operator_iata": "UA",
                "flight_number": "123",
                "origin": {"city": "Houston", "code_iata": "IAH"},
                "estimated_on": "2026-12-18T20:00:00Z",
                "cancelled": False,
            },
        ]

        arrivals = MODULE.build_arrivals(flights, date(2026, 12, 17), date(2026, 12, 18))

        self.assertEqual(
            [
                {
                    "flight": "UX 15",
                    "airline": "UX",
                    "from": "Madrid (MAD)",
                    "arrives": "12:00",
                    "thursday": True,
                    "friday": True,
                }
            ],
            arrivals,
        )

    def test_window_must_be_within_two_days(self):
        thursday = date(2026, 12, 17)
        friday = date(2026, 12, 18)
        MODULE.validate_api_window(thursday, friday, datetime(2026, 12, 17, 6, tzinfo=timezone.utc))

        with self.assertRaisesRegex(ValueError, "only available two days ahead"):
            MODULE.validate_api_window(thursday, friday, datetime(2026, 12, 17, 5, 59, tzinfo=timezone.utc))


if __name__ == "__main__":
    unittest.main()
