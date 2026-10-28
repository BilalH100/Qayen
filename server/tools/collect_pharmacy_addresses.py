"""Read public pharmacy directory pages sequentially; retain evidence, never edit DB."""
import argparse
import hashlib
import html as html_lib
import json
import re
import time
import urllib.parse
import urllib.request
import urllib.robotparser
from datetime import datetime, timezone
from pathlib import Path

BASE = "https://pharmacy.ma"
AGENT = "QayenResearch/1.0 (academic pharmacy address reconciliation)"


def parse_page(html, page_url):
    def clean(value):
        return " ".join(html_lib.unescape(re.sub(r"<[^>]*>", " ", value)).split())
    records = []
    for card in re.split(r'<div\s+class="card h-100[^\"]*">', html)[1:]:
        name = re.search(r'<h5\s+class="card-title[^\"]*">(.*?)</h5>', card, re.S)
        city = re.search(r'<a[^>]*href="/pharmacies/ville/[^\"]*"[^>]*>(.*?)</a>', card, re.S)
        link = re.search(r'<a[^>]*href="(/pharmacie/[^\"]+)"', card)
        if not (name and city and link):
            continue
        fields = {}
        for paragraph in re.findall(r'<p\b[^>]*>(.*?)</p>', card, re.S):
            label = re.search(r'<strong>(.*?)</strong>', paragraph, re.S)
            if label:
                text = clean(paragraph)
                fields[clean(label[1]).rstrip(" :")] = text.split(":", 1)[-1].strip()
        records.append({"name": clean(name[1]), "city": clean(city[1]),
                        "address": fields.get("Adresse", ""), "phone": fields.get("Téléphone", ""),
                        "url": urllib.parse.urljoin(BASE, html_lib.unescape(link[1])), "page_url": page_url})
    last = re.search(r'<a[^>]*href="([^\"]+)"[^>]*aria-label="Dernière page"', html)
    maximum = int(urllib.parse.parse_qs(urllib.parse.urlparse(html_lib.unescape(last[1])).query).get("CurrentPage", [1])[0]) if last else 1
    return records, maximum


def main():
    parser = argparse.ArgumentParser()
    parser.add_argument("--output-dir", required=True)
    args = parser.parse_args()
    directory = Path(args.output_dir)
    directory.mkdir(parents=True, exist_ok=True)
    def fetch(url):
        with urllib.request.urlopen(urllib.request.Request(url, headers={"User-Agent": AGENT}), timeout=25) as response:
            return response.read()
    robots = urllib.robotparser.RobotFileParser()
    robots.parse(fetch(BASE + "/robots.txt").decode().splitlines())
    if not robots.can_fetch(AGENT, BASE + "/pharmacies"):
        raise RuntimeError("Public listing collection is disallowed")
    all_rows, manifests = [], []
    page, last = 1, 1
    while page <= last:
        url = BASE + "/pharmacies" + (f"?CurrentPage={page}" if page > 1 else "")
        if not robots.can_fetch(AGENT, url):
            raise RuntimeError("Page disallowed by robots policy")
        cache = directory / f"pharmacy-page{page}.html"
        if cache.exists():
            data = cache.read_bytes()
        else:
            time.sleep(1)
            data = fetch(url)
            cache.write_bytes(data)
        rows, maximum = parse_page(data.decode("utf-8-sig"), url)
        if not rows or maximum > 250:
            raise RuntimeError(f"Unexpected listing structure at page {page}")
        last = max(last, maximum)
        digest = hashlib.sha256(data).hexdigest()
        captured = datetime.fromtimestamp(cache.stat().st_mtime, timezone.utc).isoformat()
        for row in rows:
            row.update({"page_sha256": digest, "retrieved_at": captured})
        all_rows.extend(rows)
        manifests.append({"page": page, "url": url, "sha256": digest, "rows": len(rows), "retrieved_at": captured})
        if page == 1 or page % 20 == 0 or page == last:
            print(f"Directory pages {page}/{last}: {len(all_rows)} source records", flush=True)
        page += 1
    output = {"provider": "Pharmacy.ma", "provider_url": BASE + "/pharmacies", "records": all_rows,
              "pages": manifests, "notes": ["Public directory evidence, not independent confirmation of current operation.",
              "Map links are address-search links; no GPS coordinates are inferred."]}
    (directory / "directory-addresses.json").write_text(json.dumps(output, ensure_ascii=False, indent=2), encoding="utf-8")
    print(f"Saved {len(all_rows)} records across {last} pages", flush=True)


if __name__ == "__main__":
    main()
