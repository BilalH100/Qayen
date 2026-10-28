const assert = require("node:assert/strict");
const { test } = require("node:test");
const fs = require("node:fs");
const path = require("node:path");
const vm = require("node:vm");
const ts = require("typescript");
const source = fs.readFileSync(path.join(__dirname, "../lib/pharmacy-location.ts"), "utf8");
const javascript = ts.transpileModule(source, { compilerOptions: { module: ts.ModuleKind.CommonJS, target: ts.ScriptTarget.ES2020 } }).outputText;
const moduleValue = { exports: {} };
vm.runInNewContext(javascript, { exports: moduleValue.exports, module: moduleValue });
const { pharmacyCoordinates } = moduleValue.exports;
const { approximatePharmacyAreas } = moduleValue.exports;

test("map pins reject missing, malformed, placeholder and out-of-range coordinates", () => {
  for (const pharmacy of [{}, { latitude: "", longitude: "" }, { latitude: " ", longitude: "-6.84" }, { latitude: "NULL", longitude: "NULL" }, { latitude: null, longitude: null }, { latitude: "34wrong", longitude: "-6.84" }, { latitude: "Infinity", longitude: "-6.84" }, { latitude: "91", longitude: "-6.84" }, { latitude: "34", longitude: "181" }, { latitude: "0", longitude: "0" }]) {
    assert.equal(pharmacyCoordinates(pharmacy), null, JSON.stringify(pharmacy));
  }
});

test("valid published pharmacy coordinates retain their exact position", () => {
  for (const [latitude, longitude] of [[34.0209, -6.8416], [30.421257, -9.598794], [35.781523, -5.811091]]) {
    const point = pharmacyCoordinates({ latitude: String(latitude), longitude: String(longitude) });
    assert.equal(point[0], latitude);
    assert.equal(point[1], longitude);
  }
});

test("city estimates group addressed pharmacies without changing their GPS", () => {
  const known = { id: 1, name: "Known", city: "Rabat - Maroc", address: "Known address", latitude: "34.02", longitude: "-6.84" };
  const missing = { id: 2, name: "Awaiting pin", city: "RABAT", address: "12 Example Street", latitude: "NULL", longitude: "NULL" };
  const noAddress = { id: 3, name: "No address", city: "Rabat", address: "", latitude: "NULL", longitude: "NULL" };
  const before = JSON.stringify([known, missing, noAddress]);
  const groups = approximatePharmacyAreas([known, missing, noAddress]);
  assert.equal(groups.length, 1);
  assert.equal(groups[0].pharmacies.length, 1);
  assert.equal(groups[0].pharmacies[0].id, 2);
  assert.equal(groups[0].point[0], 34.02);
  assert.equal(groups[0].point[1], -6.84);
  assert.equal(pharmacyCoordinates(missing), null);
  assert.equal(JSON.stringify([known, missing, noAddress]), before);
  assert.equal(approximatePharmacyAreas([missing], [known, missing])[0].point[0], 34.02);
});

test("city fallback is explicit and an unknown city is never assigned a random point", () => {
  const pending = city => ({ id: 1, name: "Awaiting pin", city, address: "Example address", latitude: "NULL", longitude: "NULL" });
  const groups = approximatePharmacyAreas([pending("MARRAKECH")]);
  assert.equal(groups.length, 1);
  assert.ok(groups[0].sourceUrl.includes("geonames.org"));
  assert.equal(pharmacyCoordinates(groups[0].pharmacies[0]), null);
  assert.equal(approximatePharmacyAreas([pending("Unknown city")]).length, 0);
});
