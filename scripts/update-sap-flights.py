#!/usr/bin/env python3
"""Refresh data/sap_flights.yaml from FlightAware AeroAPI.

Fetches SAP arrivals for Wednesday through Friday (the day before Thursday bus
through the Friday bus day). Every arrival occurrence is stored with a compact
RFC3339 UTC arrives_at timestamp. The frontend derives local date/time and bus
day eligibility from arrives_at using the America/Tegucigalpa timezone.

Each YAML entry represents a distinct (flight, origin, arrives_at) occurrence.
The bus window is derived from the wedding date in config.toml. Reruns fetch a
fresh snapshot and atomically replace the file, so stale entries are removed.
"""

import argparse
import json
import os
import re
import sys
import tempfile
import tomllib
import urllib.error
import urllib.parse
import urllib.request
from datetime import date, datetime, timezone, timedelta
from pathlib import Path
from zoneinfo import ZoneInfo


API_BASE = "https://aeroapi.flightaware.com/aeroapi"
SAP_TIMEZONE = ZoneInfo("America/Tegucigalpa")
AIRLINES = {
    "5U": "TAG Airlines",
    "9N": "Tropic Air",
    "AA": "American Airlines",
    "AM": "Aeromexico",
    "AV": "Avianca",
    "CM": "Copa Airlines",
    "DL": "Delta Air Lines",
    "H5": "CM Airlines",
    "NK": "Spirit Airlines",
    "S0": "Aerolíneas Sosa",
    "UA": "United Airlines",
    "UX": "Air Europa",
}
ORIGIN_CITIES = {
    "ATL": "Atlanta",
    "BZE": "Belize City",
    "DFW": "Dallas/Fort Worth",
    "FLL": "Fort Lauderdale",
    "GUA": "Guatemala City",
    "IAH": "Houston",
    "LCE": "La Ceiba",
    "MAD": "Madrid",
    "MEX": "Mexico City",
    "MIA": "Miami",
    "PTY": "Panama City",
    "RTB": "Roatán",
    "SAL": "San Salvador",
    "TGU": "Tegucigalpa",
}


def api_get(api_key, path, params=None):
    url = f"{API_BASE}/{path.lstrip('/')}"
    if params:
        url = f"{url}?{urllib.parse.urlencode(params)}"

    request = urllib.request.Request(
        url,
        headers={"Accept": "application/json", "x-apikey": api_key},
    )
    try:
        with urllib.request.urlopen(request, timeout=30) as response:
            return json.load(response)
    except urllib.error.HTTPError as err:
        detail = err.read().decode("utf-8", errors="replace")
        try:
            detail = json.loads(detail).get("detail", detail)
        except json.JSONDecodeError:
            pass
        raise RuntimeError(f"AeroAPI returned HTTP {err.code}: {detail}") from err
    except urllib.error.URLError as err:
        raise RuntimeError(f"AeroAPI request failed: {err.reason}") from err


def parse_args():
    parser = argparse.ArgumentParser(
        description="Refresh data/sap_flights.yaml from FlightAware AeroAPI.",
    )
    parser.add_argument(
        "--output",
        type=Path,
        default=Path(__file__).resolve().parents[1] / "data" / "sap_flights.yaml",
        help="Output YAML path.",
    )
    return parser.parse_args()


def arrival_window():
    config_path = Path(__file__).resolve().parents[1] / "config.toml"
    try:
        with config_path.open("rb") as config_file:
            wedding = datetime.fromisoformat(tomllib.load(config_file)["params"]["weddingDate"]).date()
    except (KeyError, TypeError, ValueError, OSError) as err:
        raise ValueError(f"Could not read params.weddingDate from {config_path}") from err

    wednesday = wedding - timedelta(days=3)
    friday = wedding - timedelta(days=1)
    if wednesday.weekday() != 2 or friday.weekday() != 4:
        raise ValueError("params.weddingDate must be a Saturday")
    return wednesday, friday


def flight_label(flight):
    operator = flight.get("operator_iata")
    number = flight.get("flight_number")
    if operator and number:
        return f"{operator.upper()}{number.upper()}"

    ident = flight.get("ident_iata") or flight.get("ident")
    if not ident:
        return ""
    ident = ident.replace(" ", "").upper()
    match = re.fullmatch(r"([A-Z0-9]{2})(\d+[A-Z]?)", ident)
    return f"{match.group(1)}{match.group(2)}" if match else ident


def scheduled_arrival_utc(flight):
    """Return the scheduled arrival as a compact RFC3339 UTC string, or None."""
    value = flight.get("scheduled_in") or flight.get("scheduled_on")
    if not value:
        return None
    dt = datetime.fromisoformat(value.replace("Z", "+00:00")).astimezone(timezone.utc)
    return dt.strftime("%Y-%m-%dT%H:%MZ")


def fetch_schedules(api_key, wednesday, friday):
    """Fetch SAP arrivals from Wednesday through Friday (inclusive)."""
    params = {
        "destination": "SAP",
        "include_codeshares": "false",
        "include_regional": "true",
        "max_pages": 10,
    }
    end = friday + timedelta(days=1)
    payload = api_get(api_key, f"schedules/{wednesday.isoformat()}/{end.isoformat()}", params)
    if (payload.get("links") or {}).get("next"):
        raise RuntimeError("AeroAPI returned more than 10 pages; output unchanged")
    return payload.get("scheduled", [])


def build_arrivals(flights, wednesday, friday):
    """Build one YAML entry per distinct (flight, origin, arrives_at) occurrence.

    Arrivals on Wed/Thu/Fri local SAP time are included. The arrives_at UTC
    timestamp is the deduplication key — same flight can appear multiple days.
    """
    arrivals = {}

    for flight in flights:
        arrives_at = scheduled_arrival_utc(flight)
        if not arrives_at:
            continue

        # Convert UTC back to local to check it falls in the fetch window
        dt_utc = datetime.fromisoformat(arrives_at.replace("Z", "+00:00"))
        local_date = dt_utc.astimezone(SAP_TIMEZONE).date()
        if not wednesday <= local_date <= friday:
            continue

        label = flight_label(flight)
        origin_code = (
            flight.get("origin_iata")
            or flight.get("origin_icao")
            or flight.get("origin")
            or ""
        )
        if not label or not origin_code:
            continue

        match = re.match(r"^([A-Z0-9]{2})", label.replace(" ", ""))
        airline_code = flight.get("operator_iata") or (match.group(1) if match else "")
        airline = AIRLINES.get(airline_code, airline_code)
        city = ORIGIN_CITIES.get(origin_code, origin_code)

        # Deduplicate by (flight, origin, timestamp) — preserves same-day
        # duplicate codeshares while keeping distinct days for recurring flights.
        key = (label, origin_code, arrives_at)
        if key in arrivals:
            continue

        arrivals[key] = {
            "flight": label,
            "airline": airline,
            "from": f"{city} ({origin_code})",
            "arrives_at": arrives_at,
        }

    return sorted(arrivals.values(), key=lambda item: item["arrives_at"])


def yaml_quote(value):
    return json.dumps(value, ensure_ascii=False)


def write_yaml(path, arrivals):
    lines = ["arrivals:"]
    for flight in arrivals:
        lines.extend(
            [
                f"  - flight: {yaml_quote(flight['flight'])}",
                f"    airline: {yaml_quote(flight['airline'])}",
                f"    from: {yaml_quote(flight['from'])}",
                f"    arrives_at: {yaml_quote(flight['arrives_at'])}",
                "",
            ]
        )

    path.parent.mkdir(parents=True, exist_ok=True)
    with tempfile.NamedTemporaryFile("w", encoding="utf-8", dir=path.parent, delete=False) as output:
        output.write("\n".join(lines))
        temporary_path = output.name
    os.replace(temporary_path, path)


def main():
    args = parse_args()
    wednesday, friday = arrival_window()
    if friday > date.today() + timedelta(days=365):
        raise ValueError("AeroAPI schedules are only available up to one year ahead")

    api_key = os.environ.get("AEROAPI_KEY")
    if not api_key:
        raise ValueError("AEROAPI_KEY is required")

    flights = fetch_schedules(api_key, wednesday, friday)
    arrivals = build_arrivals(flights, wednesday, friday)
    if not arrivals:
        raise RuntimeError("AeroAPI returned no SAP arrivals for the bus window; output unchanged")

    write_yaml(args.output, arrivals)
    print(f"Wrote {len(arrivals)} arrivals to {args.output}")


if __name__ == "__main__":
    try:
        main()
    except (RuntimeError, ValueError) as err:
        print(f"error: {err}", file=sys.stderr)
        sys.exit(1)
