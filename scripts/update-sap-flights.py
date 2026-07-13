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
    ident = flight.get("ident_iata") or flight.get("actual_ident_iata") or flight.get("ident")
    if not ident:
        return ""
    ident = ident.replace(" ", "").upper()
    match = re.fullmatch(r"([A-Z0-9]{2,3})(\d+[A-Z]?)", ident)
    return f"{match.group(1)} {match.group(2)}" if match else ident


def operator_code(flight):
    ident = flight.get("actual_ident_iata") or flight.get("ident_iata") or ""
    match = re.match(r"^([A-Z0-9]{2})\d", ident.replace(" ", "").upper())
    return match.group(1) if match else ""


def scheduled_arrival(flight):
    value = flight.get("scheduled_in") or flight.get("scheduled_on")
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
    path = f"schedules/{thursday.isoformat()}/{(friday + timedelta(days=1)).isoformat()}"
    flights = []

    while path:
        payload = api_get(api_key, path, params)
        flights.extend(payload.get("scheduled", []))
        next_url = (payload.get("links") or {}).get("next")
        path = next_url or ""
        params = None

    return flights


def metadata_name(api_key, resource, code, field, cache):
    if not code:
        return ""
    key = (resource, code)
    if key not in cache:
        try:
            cache[key] = api_get(api_key, f"{resource}/{urllib.parse.quote(code)}")
        except RuntimeError as err:
            print(f"warning: could not resolve {resource} {code}: {err}", file=sys.stderr)
            cache[key] = {}
    return cache[key].get(field) or ""


def build_arrivals(api_key, flights, thursday, friday):
    cache = {}
    arrivals = {}

    for flight in flights:
        arrival = scheduled_arrival(flight)
        if not arrival or arrival.date() not in (thursday, friday) or arrival.time() > PICKUP_CUTOFF:
            continue

        label = flight_label(flight)
        origin_ref = flight.get("origin_iata") or flight.get("origin") or ""
        if not label or not origin_ref:
            continue

        city = metadata_name(api_key, "airports", origin_ref, "city", cache) or origin_ref
        origin = metadata_name(api_key, "airports", origin_ref, "code_iata", cache) or origin_ref
        code = operator_code(flight)
        airline = metadata_name(api_key, "operators", code, "name", cache) or code
        key = (label, airline, city, origin, arrival.strftime("%H:%M"))
        entry = arrivals.setdefault(
            key,
            {
                "flight": label,
                "airline": airline,
                "from": f"{city} ({origin})",
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
    api_key = os.environ.get("AEROAPI_KEY")
    if not api_key:
        raise ValueError("AEROAPI_KEY is required")

    thursday = parse_bus_date(args.thursday, 3, "Thursday")
    friday = parse_bus_date(args.friday, 4, "Friday")
    if friday <= thursday or friday - thursday > timedelta(days=21):
        raise ValueError("Friday must follow Thursday within AeroAPI's three-week schedule window")
    if friday > date.today() + timedelta(days=365):
        raise ValueError("AeroAPI schedules are only available up to one year ahead")

    flights = fetch_schedules(api_key, thursday, friday)
    arrivals = build_arrivals(api_key, flights, thursday, friday)
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
