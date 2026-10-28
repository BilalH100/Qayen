const assert = require("node:assert/strict");
const { test } = require("node:test");
const fs = require("node:fs");
const path = require("node:path");
const vm = require("node:vm");
const ts = require("typescript");
const source = fs.readFileSync(path.join(__dirname, "../lib/platform-search.ts"), "utf8");
const javascript = ts.transpileModule(source, { compilerOptions: { module: ts.ModuleKind.CommonJS, target: ts.ScriptTarget.ES2020 } }).outputText;
const moduleValue = { exports: {} };
vm.runInNewContext(javascript, { exports: moduleValue.exports, module: moduleValue });
const { searchPlatform, searchableSections } = moduleValue.exports;
const medication = { id: "medication-1", title: "Doliprane", description: "Paracétamol · 500 mg", category: "Medications", href: "/medications/1" };
const pharmacy = { id: "pharmacy-2", title: "Pharmacie de la Paix", description: "12 Avenue Hassan · Rabat", category: "Pharmacies", href: "/pharmacies/2" };
const items = [...searchableSections, medication, pharmacy];

test("general search finds medications by ingredient and pharmacies by city/address", () => {
  assert.equal(searchPlatform(items, "PARACETAMOL 500")[0].href, medication.href);
  assert.equal(searchPlatform(items, "paix rabat")[0].href, pharmacy.href);
  assert.equal(searchPlatform(items, "Hassan")[0].href, pharmacy.href);
  assert.equal(searchPlatform(items, "Doliprane Rabat").length, 0);
});
test("section discovery supports French, Arabic and unfinished catalogs", () => {
  for (const [query, href] of [["médicaments", "/medications"], ["الأَدْوِيَة", "/medications"], ["para", "/parapharmacies"], ["matériel médical", "/medical-materials"], ["accessoires", "/supplies-accessories"], ["اتصل بنا", "/contact"], ["about", "/about"]]) {
    assert.ok(searchPlatform(items, query).some(result => result.href === href), query);
  }
  assert.equal(searchPlatform(items, "para").find(result => result.href === "/parapharmacies").description, "Coming soon");
});
test("empty and punctuation-only queries return no results, and exact titles rank first", () => {
  for (const query of ["", "  ", "---"]) assert.equal(searchPlatform(items, query).length, 0);
  assert.equal(searchPlatform(items, "Doliprane")[0].id, medication.id);
  assert.equal(searchPlatform(items, "nonexistentxyz").length, 0);
});
