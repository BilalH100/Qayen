"""Collect published pharmacy map points, without geocoding streets/city centres.

Only public profiles named in the site's sitemap and matching an addressed
database pharmacy are read. Cache responses, use one connection and <=2.5 rps.
"""
import argparse
import hashlib
import http.client
import json
import math
import re
import ssl
import time
import unicodedata
import urllib.parse
import urllib.robotparser
import xml.etree.ElementTree as ET
from collections import defaultdict
from datetime import datetime, timezone
from html.parser import HTMLParser
from pathlib import Path

BASE = "https://pharmap.ma"
AGENT = "QayenResearch/1.0 (one-time academic pharmacy location reconciliation)"


def name_key(value):
    value = "".join(c for c in unicodedata.normalize("NFD", value.upper()) if not unicodedata.combining(c))
    return "".join(c for c in re.sub(r"^PHARMACIE\s+", "", value) if c.isalnum())


class StructuredData(HTMLParser):
    def __init__(self):
        super().__init__()
        self.active = False
        self.parts = []
        self.documents = []

    def handle_starttag(self, tag, attrs):
        if tag == "script" and dict(attrs).get("type") == "application/ld+json":
            self.active = True
            self.parts = []

    def handle_data(self, data):
        if self.active:
            self.parts.append(data)

    def handle_endtag(self, tag):
        if tag == "script" and self.active:
            self.documents.append(json.loads("".join(self.parts)))
            self.active = False


def parse_profiles(html, page_url, page_hash, captured):
    parser = StructuredData()
    parser.feed(html)
    result = []
    def walk(item):
        if isinstance(item, list):
            for child in item:
                walk(child)
        elif isinstance(item, dict):
            kinds = item.get("@type", [])
            if kinds == "Pharmacy" or isinstance(kinds, list) and "Pharmacy" in kinds:
                geo = item.get("geo") or {}
                address = item.get("address") or {}
                url = item.get("url") or page_url
                if not url.startswith(BASE + "/pharmacie/"):
                    return
                try:
                    lat, lon = float(geo["latitude"]), float(geo["longitude"])
                except (KeyError, ValueError, TypeError):
                    lat = lon = None
                if lat is not None and not (math.isfinite(lat) and math.isfinite(lon)):
                    lat = lon = None
                result.append({"name": item.get("name", ""), "city": address.get("addressLocality", ""),
                               "address": address.get("streetAddress", ""), "phone": item.get("telephone", ""),
                               "latitude": lat, "longitude": lon, "source_url": url, "page_url": page_url,
                               "page_sha256": page_hash, "retrieved_at": captured})
            if "@graph" in item:
                walk(item["@graph"])
    for document in parser.documents:
        walk(document)
    return result


def main():
    args = argparse.ArgumentParser()
    args.add_argument("--directory", required=True)
    args.add_argument("--pharmacies", required=True)
    options = args.parse_args()
    directory = Path(options.directory)
    directory.mkdir(parents=True, exist_ok=True)
    cache = directory / "profiles"
    cache.mkdir(exist_ok=True)
    pharmacies = json.loads(Path(options.pharmacies).read_text(encoding="utf-8-sig"))["pharmacies"]
    def missing(p):
        try:
            return not (math.isfinite(float(p["latitude"])) and math.isfinite(float(p["longitude"])))
        except (KeyError, ValueError, TypeError):
            return True
    names = {name_key(p["name"]) for p in pharmacies if p.get("address", "").strip() and missing(p)}
    robot = urllib.robotparser.RobotFileParser()
    robot.parse((directory / "pharmap-robots.txt").read_text().splitlines())
    sitemap = ET.parse(directory / "pharmap-sitemap.xml")
    urls = set()
    for element in sitemap.getroot().iter():
        value = element.text or ""
        if element.tag.endswith("loc") and value.startswith(BASE + "/pharmacie/"):
            slug = re.sub(r"-\d+$", "", value.rsplit("/", 1)[-1]).replace("-", " ")
            if name_key(slug) in names:
                urls.add(value)
    urls = sorted(urls)
    connection = http.client.HTTPSConnection("pharmap.ma", timeout=15, context=ssl.create_default_context())
    last_fetch = 0.0
    records, pages, failures = {}, [], []
    output = directory / "published-locations.json"
    def save():
        bundle = {"provider": "pharmap.ma", "records": list(records.values()), "pages": pages,
                  "failures": failures, "candidate_profiles": len(urls), "processed_profiles": len(pages) + len(failures),
                  "complete": len(pages) + len(failures) == len(urls),
                  "notes": ["Published pharmacy POI coordinates, not street/city geocoding.",
                            "Current operation and on-site pin accuracy are not independently verified."]}
        output.write_text(json.dumps(bundle, ensure_ascii=False, indent=2), encoding="utf-8")
    for index, url in enumerate(urls, 1):
        if not robot.can_fetch(AGENT, url):
            failures.append({"url": url, "reason": "robots_disallowed"})
            continue
        path = cache / (url.rsplit("/", 1)[-1] + ".html")
        try:
            if not path.exists():
                time.sleep(max(0, 0.4 - (time.monotonic() - last_fetch)))
                last_fetch = time.monotonic()
                connection.request("GET", urllib.parse.quote(urllib.parse.urlparse(url).path, safe="/%-._~"),
                                   headers={"User-Agent": AGENT, "Accept": "text/html"})
                response = connection.getresponse()
                data = response.read()
                if response.status in (403, 429):
                    save()
                    raise SystemExit(f"Collection stopped: HTTP {response.status}; do not retry automatically")
                if response.status != 200:
                    raise ValueError(f"HTTP {response.status}")
                path.write_bytes(data)
            data = path.read_bytes()
            captured = datetime.fromtimestamp(path.stat().st_mtime, timezone.utc).isoformat()
            digest = hashlib.sha256(data).hexdigest()
            profiles = parse_profiles(data.decode("utf-8-sig"), url, digest, captured)
            if not profiles:
                raise ValueError("no published pharmacy structured data")
            for record in profiles:
                # The canonical public profile is the source identifier.
                records[record["source_url"]] = record
            pages.append({"url": url, "sha256": digest, "records": len(profiles), "retrieved_at": captured})
        except (OSError, ValueError, http.client.HTTPException) as exc:
            failures.append({"url": url, "reason": type(exc).__name__ + ": " + str(exc)})
            connection.close()
            connection = http.client.HTTPSConnection("pharmap.ma", timeout=15, context=ssl.create_default_context())
        if index == 1 or index % 100 == 0 or index == len(urls):
            save()
            print(f"Profiles {index}/{len(urls)}: {len(records)} evidence records, {len(failures)} unavailable", flush=True)
    save()
    connection.close()


if __name__ == "__main__":
    main()
