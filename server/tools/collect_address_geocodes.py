"""One-time, cached pharmacy-address lookup; never accept a city/street centre.

Reads public business addresses only. Single connection, <=1 request/second;
stops on throttling/access denial. Raw results are retained for strict matching.
"""
import argparse
import hashlib
import http.client
import json
import math
import ssl
import time
import unicodedata
import urllib.parse
from datetime import datetime, timezone
from pathlib import Path

AGENT = "QayenResearch/1.0 (one-time public pharmacy address reconciliation)"


def latin_name(value):
    # OSM often appends the Arabic spelling to its Latin pharmacy name.
    return " ".join("".join(c if c.isdigit() or not c.isalpha() or "LATIN" in unicodedata.name(c, "") else " " for c in value).split())


def main():
    parser = argparse.ArgumentParser()
    parser.add_argument("--preview", required=True)
    parser.add_argument("--directory", required=True)
    args = parser.parse_args()
    directory = Path(args.directory)
    cache = directory / "geocoder-cache"
    cache.mkdir(parents=True, exist_ok=True)
    preview = json.loads(Path(args.preview).read_text(encoding="utf-8"))
    targets = [r for r in preview["decisions"] if r["address"] and r["status"] not in ("preserve_existing_coordinates", "fill_missing_coordinates")]
    connection = http.client.HTTPSConnection("photon.komoot.io", timeout=20, context=ssl.create_default_context())
    last = 0.0
    records, pages, failures = {}, [], []
    def save(complete=False):
        bundle = {"provider": "OpenStreetMap/Photon", "records": list(records.values()), "complete": complete,
                  "targets": len(targets), "processed": len(pages) + len(failures), "pages": pages, "failures": failures,
                  "notes": ["Only named OSM pharmacy objects are collected; roads, localities and buildings are excluded.",
                            "A name/city/street or building match must still be established before saving a pin."]}
        (directory / "osm-pharmacy-locations.json").write_text(json.dumps(bundle, ensure_ascii=False, indent=2), encoding="utf-8")
    for index, row in enumerate(targets, 1):
        def query_url(query):
            return "https://photon.komoot.io/api/?" + urllib.parse.urlencode({"q": query, "limit": 5, "lang": "fr", "bbox": "-18,20,-0.8,36.1", "osm_tag": "amenity:pharmacy"})
        full_url = query_url(f"{row['name']}, {row['address']}, {row['city']}, Maroc")
        full_cache = cache / (hashlib.sha256(full_url.encode()).hexdigest() + ".json")
        # Sparse map records often omit address details. Search name/city,
        # then check the returned street independently; never trust rank alone.
        url = query_url(f"{row['name']}, {row['city']}, Maroc")
        if full_cache.exists() and json.loads(full_cache.read_bytes()).get("features"):
            url = full_url
        path = cache / (hashlib.sha256(url.encode()).hexdigest() + ".json")
        try:
            if not path.exists():
                time.sleep(max(0, 1.0 - (time.monotonic() - last)))
                last = time.monotonic()
                parsed = urllib.parse.urlsplit(url)
                connection.request("GET", parsed.path + "?" + parsed.query, headers={"User-Agent": AGENT, "Accept": "application/json"})
                response = connection.getresponse()
                data = response.read()
                if response.status in (403, 429):
                    save()
                    raise SystemExit(f"Stopped on HTTP {response.status}; no automatic retry")
                if response.status != 200:
                    raise ValueError(f"HTTP {response.status}")
                json.loads(data)
                path.write_bytes(data)
            data = path.read_bytes()
            captured = datetime.fromtimestamp(path.stat().st_mtime, timezone.utc).isoformat()
            digest = hashlib.sha256(data).hexdigest()
            features = json.loads(data)["features"]
            accepted = 0
            for item in features:
                properties, geometry = item.get("properties", {}), item.get("geometry", {})
                if properties.get("osm_key") != "amenity" or properties.get("osm_value") != "pharmacy" or geometry.get("type") != "Point":
                    continue
                city, street = properties.get("city", ""), properties.get("street", "")
                name = properties.get("name", "")
                osm_type, osm_id = properties.get("osm_type"), properties.get("osm_id")
                types = {"N": "node", "W": "way", "R": "relation"}
                if not (name and city and street and osm_type in types and isinstance(osm_id, int)):
                    continue
                lon, lat = geometry["coordinates"]
                if not (math.isfinite(lat) and math.isfinite(lon) and 20 <= lat <= 36.1 and -18 <= lon <= -0.8):
                    continue
                source_url = f"https://www.openstreetmap.org/{types[osm_type]}/{osm_id}"
                record = {"provider": "OpenStreetMap/Photon", "kind": "pharmacy_poi", "name": latin_name(name),
                          "source_name": name, "city": city, "address": " ".join(str(x) for x in (properties.get("housenumber", ""), street, city) if x),
                          "street": street, "housenumber": properties.get("housenumber", ""), "phone": "",
                          "latitude": lat, "longitude": lon, "source_url": source_url, "page_url": url,
                          "page_sha256": digest, "retrieved_at": captured, "osm_key": "amenity", "osm_value": "pharmacy"}
                if record["name"]:
                    records.setdefault(source_url, record)
                    accepted += 1
            pages.append({"pharmacy_id": row["pharmacy_id"], "url": url, "sha256": digest, "pharmacy_objects": accepted, "retrieved_at": captured})
        except (OSError, ValueError, KeyError, http.client.HTTPException) as exc:
            failures.append({"pharmacy_id": row["pharmacy_id"], "reason": str(exc)})
            connection.close()
            connection = http.client.HTTPSConnection("photon.komoot.io", timeout=20, context=ssl.create_default_context())
            if len(failures) >= 3 and not pages:
                save()
                raise SystemExit("Service unavailable; stopped after three initial failures")
        if index == 1 or index % 50 == 0 or index == len(targets):
            save(index == len(targets))
            print(f"Address lookups {index}/{len(targets)}: {len(records)} named pharmacy objects, {len(failures)} unavailable", flush=True)
    save(True)
    connection.close()


if __name__ == "__main__":
    main()
