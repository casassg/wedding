import importlib.util
import unittest
from datetime import date
from pathlib import Path
from unittest import mock


SCRIPT = Path(__file__).with_name("update-sap-flights.py")
SPEC = importlib.util.spec_from_file_location("update_sap_flights", SCRIPT)
MODULE = importlib.util.module_from_spec(SPEC)
SPEC.loader.exec_module(MODULE)


class UpdateSapFlightsTest(unittest.TestCase):
    def test_fetch_uses_one_schedules_request(self):
        response = {"links": {"next": None}, "scheduled": [{"ident": "UAL123"}]}

        with mock.patch.object(MODULE, "api_get", return_value=response) as api_get:
            flights = MODULE.fetch_schedules("key", date(2026, 12, 17), date(2026, 12, 18))

        self.assertEqual([{"ident": "UAL123"}], flights)
        api_get.assert_called_once_with(
            "key",
            "schedules/2026-12-17/2026-12-19",
            {
                "destination": "SAP",
                "include_codeshares": "false",
                "include_regional": "true",
                "max_pages": 10,
            },
        )

    def test_fetch_refuses_incomplete_response(self):
        response = {"links": {"next": "/more"}, "scheduled": []}
        with mock.patch.object(MODULE, "api_get", return_value=response):
            with self.assertRaisesRegex(RuntimeError, "more than 10 pages"):
                MODULE.fetch_schedules("key", date(2026, 12, 17), date(2026, 12, 18))

    def test_build_arrivals_uses_local_names_and_combines_days(self):
        flights = [
            {
                "ident_iata": "UX15",
                "origin_iata": "MAD",
                "scheduled_in": "2026-12-17T18:00:00Z",
            },
            {
                "ident_iata": "UX15",
                "origin_iata": "MAD",
                "scheduled_in": "2026-12-18T18:00:00Z",
            },
            {
                "ident_iata": "UA123",
                "origin_iata": "IAH",
                "scheduled_in": "2026-12-18T20:00:00Z",
            },
        ]

        arrivals = MODULE.build_arrivals(flights, date(2026, 12, 17), date(2026, 12, 18))

        self.assertEqual(
            [
                {
                    "flight": "UX 15",
                    "airline": "Air Europa",
                    "from": "Madrid (MAD)",
                    "arrives": "12:00",
                    "thursday": True,
                    "friday": True,
                }
            ],
            arrivals,
        )

    def test_unknown_codes_fall_back_without_more_requests(self):
        arrivals = MODULE.build_arrivals(
            [{"ident_iata": "ZZ42", "origin_iata": "XYZ", "scheduled_in": "2026-12-17T18:00:00Z"}],
            date(2026, 12, 17),
            date(2026, 12, 18),
        )
        self.assertEqual("ZZ", arrivals[0]["airline"])
        self.assertEqual("XYZ (XYZ)", arrivals[0]["from"])


if __name__ == "__main__":
    unittest.main()
