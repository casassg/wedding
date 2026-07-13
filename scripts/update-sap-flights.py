#!/usr/bin/env python3

import argparse
import json
import os
import re
import sys
import tempfile
import urllib.error
import urllib.parse
import urllib.request
from datetime import date, datetime, time, timedelta
from pathlib import Path
from zoneinfo import ZoneInfo


API_BASE = "https://aeroapi.flightaware.com/aeroapi"
SAP_TIMEZONE = ZoneInfo("America/Tegucigalpa")
PICKUP_CUTOFF = time(13, 0)
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
    if path.startswith("http"):
        url = path
    elif path.startswith("/aeroapi/"):
        url = f"https://aeroapi.flightaware.com{path}"
    else:
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
    parser.add_argument("--thursday", default="2026-12-17", help="Thursday bus date (YYYY-MM-DD).")
    parser.add_argument("--friday", default="2026-12-18", help="Friday bus date (YYYY-MM-DD).")
    parser.add_argument(
        "--output",
        type=Path,
        default=Path(__file__).resolve().parents[1] / "data" / "sap_flights.yaml",
        help="Output YAML path.",
    )
    return parser.parse_args()


def parse_bus_date(value, weekday, label):
    try:
        parsed = date.fromisoformat(value)
    except ValueError as err:
        raise ValueError(f"{label} must use YYYY-MM-DD") from err
    if parsed.weekday() != weekday:
        raise ValueError(f"{label} date {parsed} is not a {label}")
    return parsed


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


def scheduled_arrival(flight):
    value = (
        flight.get("scheduled_in")
        or flight.get("scheduled_on")
    )
    if not value:
        return None
    return datetime.fromisoformat(value.replace("Z", "+00:00")).astimezone(SAP_TIMEZONE)


def fetch_schedules(api_key, thursday, friday):
    params = {
        "destination": "SAP",
        "include_codeshares": "false",
        "include_regional": "true",
        "max_pages": 10,
    }
    end = friday + timedelta(days=1)
    payload = api_get(api_key, f"schedules/{thursday.isoformat()}/{end.isoformat()}", params)
    if (payload.get("links") or {}).get("next"):
        raise RuntimeError("AeroAPI returned more than 10 pages; output unchanged")
    return payload.get("scheduled", [])


def build_arrivals(flights, thursday, friday):
    arrivals = {}

    for flight in flights:
        arrival = scheduled_arrival(flight)
        if not arrival or arrival.date() not in (thursday, friday):
            continue

        works_for_thursday = arrival.date() == thursday and arrival.time() <= PICKUP_CUTOFF
        works_for_friday = arrival.date() == thursday or arrival.time() <= PICKUP_CUTOFF
        if not works_for_thursday and not works_for_friday:
            continue

        label = flight_label(flight)
        origin_code = flight.get("origin_iata") or flight.get("origin_icao") or flight.get("origin") or ""
        if not label or not origin_code:
            continue

        match = re.match(r"^([A-Z0-9]{2})", label.replace(" ", ""))
        airline_code = flight.get("operator_iata") or (match.group(1) if match else "")
        airline = AIRLINES.get(airline_code, airline_code)
        city = ORIGIN_CITIES.get(origin_code, origin_code)
        key = (label, airline, city, origin_code, arrival.strftime("%H:%M"))
        entry = arrivals.setdefault(
            key,
            {
                "flight": label,
                "airline": airline,
                "from": f"{city} ({origin_code})",
                "arrives": arrival.strftime("%H:%M"),
                "thursday": False,
                "friday": False,
            },
        )
        entry["thursday"] = entry["thursday"] or works_for_thursday
        entry["friday"] = entry["friday"] or works_for_friday

    return sorted(arrivals.values(), key=lambda item: (item["arrives"], item["flight"], item["from"]))


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
                f"    arrives: {yaml_quote(flight['arrives'])}",
                f"    thursday: {str(flight['thursday']).lower()}",
                f"    friday: {str(flight['friday']).lower()}",
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
    thursday = parse_bus_date(args.thursday, 3, "Thursday")
    friday = parse_bus_date(args.friday, 4, "Friday")
    if friday <= thursday or friday - thursday > timedelta(days=21):
        raise ValueError("Friday must follow Thursday")
    if friday > date.today() + timedelta(days=365):
        raise ValueError("AeroAPI schedules are only available up to one year ahead")

    api_key = os.environ.get("AEROAPI_KEY")
    if not api_key:
        raise ValueError("AEROAPI_KEY is required")

    flights = fetch_schedules(api_key, thursday, friday)
    arrivals = build_arrivals(flights, thursday, friday)
    if not arrivals:
        raise RuntimeError("AeroAPI returned no SAP arrivals compatible with the bus schedule; output unchanged")

    write_yaml(args.output, arrivals)
    print(f"Wrote {len(arrivals)} arrivals to {args.output}")


if __name__ == "__main__":
    try:
        main()
    except (RuntimeError, ValueError) as err:
        print(f"error: {err}", file=sys.stderr)
        sys.exit(1)
