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
from datetime import date, datetime, time, timedelta, timezone
from pathlib import Path
from zoneinfo import ZoneInfo


API_BASE = "https://aeroapi.flightaware.com/aeroapi"
SAP_TIMEZONE = ZoneInfo("America/Tegucigalpa")
PICKUP_CUTOFF = time(13, 0)


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
        return f"{operator.upper()} {number.upper()}"

    ident = flight.get("ident_iata") or flight.get("ident")
    if not ident:
        return ""
    ident = ident.replace(" ", "").upper()
    match = re.fullmatch(r"([A-Z0-9]{2})(\d+[A-Z]?)", ident)
    return f"{match.group(1)} {match.group(2)}" if match else ident


def scheduled_arrival(flight):
    value = (
        flight.get("estimated_on")
        or flight.get("scheduled_on")
        or flight.get("estimated_in")
        or flight.get("scheduled_in")
    )
    if not value:
        return None
    return datetime.fromisoformat(value.replace("Z", "+00:00")).astimezone(SAP_TIMEZONE)


def flight_window(thursday, friday):
    start = datetime.combine(thursday, time.min, tzinfo=SAP_TIMEZONE).astimezone(timezone.utc)
    end = datetime.combine(friday + timedelta(days=1), time.min, tzinfo=SAP_TIMEZONE).astimezone(timezone.utc)
    return start, end


def validate_api_window(thursday, friday, now=None):
    start, end = flight_window(thursday, friday)
    now = now or datetime.now(timezone.utc)
    if start < now - timedelta(days=10):
        raise ValueError("AeroAPI scheduled arrivals are only available for the last 10 days")
    if end > now + timedelta(days=2):
        raise ValueError(
            "AeroAPI scheduled arrivals are only available two days ahead; "
            f"refresh {thursday}–{friday} on or after midnight in Honduras on {thursday}"
        )
    return start, end


def fetch_scheduled_arrivals(api_key, start, end):
    params = {
        "start": start.isoformat().replace("+00:00", "Z"),
        "end": end.isoformat().replace("+00:00", "Z"),
        "type": "Airline",
        "max_pages": 10,
    }
    payload = api_get(api_key, "airports/SAP/flights/scheduled_arrivals", params)
    if (payload.get("links") or {}).get("next"):
        raise RuntimeError("AeroAPI returned more than 10 pages; output unchanged")
    return payload.get("scheduled_arrivals", [])


def build_arrivals(flights, thursday, friday):
    arrivals = {}

    for flight in flights:
        if flight.get("cancelled"):
            continue
        arrival = scheduled_arrival(flight)
        if not arrival or arrival.date() not in (thursday, friday) or arrival.time() > PICKUP_CUTOFF:
            continue

        label = flight_label(flight)
        origin = flight.get("origin") or {}
        origin_code = origin.get("code_iata") or origin.get("code_icao") or origin.get("code") or ""
        if not label or not origin_code:
            continue

        city = origin.get("city") or origin.get("name") or origin_code
        airline = flight.get("operator_iata") or flight.get("operator_icao") or flight.get("operator") or ""
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
        entry["thursday" if arrival.date() == thursday else "friday"] = True

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
    start, end = validate_api_window(thursday, friday)

    api_key = os.environ.get("AEROAPI_KEY")
    if not api_key:
        raise ValueError("AEROAPI_KEY is required")

    flights = fetch_scheduled_arrivals(api_key, start, end)
    arrivals = build_arrivals(flights, thursday, friday)
    if not arrivals:
        raise RuntimeError("AeroAPI returned no SAP arrivals before the 1:00 PM pickup cutoff; output unchanged")

    write_yaml(args.output, arrivals)
    print(f"Wrote {len(arrivals)} arrivals to {args.output}")


if __name__ == "__main__":
    try:
        main()
    except (RuntimeError, ValueError) as err:
        print(f"error: {err}", file=sys.stderr)
        sys.exit(1)
