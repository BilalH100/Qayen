const assert = require("node:assert/strict");
const fs = require("node:fs");
const path = require("node:path");
const vm = require("node:vm");
const { test } = require("node:test");
const ts = require("typescript");

const webRoot = path.resolve(__dirname, "..");

function loadTs(relativePath, globals = {}) {
  const file = path.join(webRoot, relativePath);
  const source = ts.transpileModule(fs.readFileSync(file, "utf8"), {
    compilerOptions: { module: ts.ModuleKind.CommonJS, target: ts.ScriptTarget.ES2020 },
  }).outputText;
  const module = { exports: {} };
  const requireModule = (name) => {
    if (name === "./translation-copy" || name === "@/lib/translation-copy") return copy;
    if (name === "next/server") return { NextResponse: { json: Response.json } };
    throw new Error(`Unexpected import: ${name}`);
  };
  vm.runInNewContext(source, { module, exports: module.exports, require: requireModule, process, AbortController, AbortSignal, ...globals }, { filename: file });
  return module.exports;
}

const copy = loadTs("lib/translation-copy.ts");

test("all curated interface copy has French and Arabic text, with no Latin UI words in Arabic", () => {
  assert.ok(copy.UI_TRANSLATIONS.length > 400);
  for (const [source, french, arabic] of copy.UI_TRANSLATIONS) {
    assert.ok(french.trim(), source);
    assert.match(arabic, /[\u0600-\u06ff]/, source);
    assert.doesNotMatch(arabic, /[A-Za-z]/, source);
    assert.ok(copy.getUiTranslation(source, "fr"), source);
    assert.ok(copy.getUiTranslation(source, "ar"), source);
  }
});

test("labels handle whitespace, case, punctuation, French source text and contact values", () => {
  assert.equal(copy.getUiTranslation("  Quantity   needed: ", "ar"), "الكمية المطلوبة:");
  assert.equal(copy.getUiTranslation("SEARCHING…", "ar"), "جارٍ البحث...");
  assert.equal(copy.getUiTranslation("Médicaments", "ar"), "الأدوية");
  assert.equal(copy.getUiTranslation("Email: info@Kayena.com", "ar"), "البريد الإلكتروني: info@Kayena.com");
  assert.equal(copy.getUiTranslation("Phone: (123) 456-7890", "fr"), "Téléphone: (123) 456-7890");
  assert.equal(copy.getUiTranslation("0.8 miles", "ar"), "0.8 ميل");
  assert.equal(copy.getUiTranslation("9:00 PM", "ar"), "9:00 م");
  assert.equal(copy.getUiTranslation("9:00 PM", "fr"), "21:00");
  assert.equal(copy.getUiTranslation("Unknown pharmacy description", "ar"), undefined);
});

test("dynamic notifications translate the sentence while preserving names and quantities", () => {
  assert.equal(copy.getUiTranslation("Welcome back, Amal!", "ar"), "مرحبًا بعودتك، Amal!");
  assert.equal(copy.getUiTranslation("Central (ID: 42) has it in stock.", "ar"), "Central (الرقم: 42) يتوفر لديها هذا الدواء في المخزون.");
  assert.equal(copy.getUiTranslation("Please choose between 1 and 12.", "ar"), "يرجى اختيار كمية بين 1 و12.");
  assert.equal(copy.getUiTranslation("2 items reserved. Remaining pharmacy stock: 8.", "fr"), "2 unité(s) réservée(s). Stock restant en pharmacie : 8.");
  assert.equal(copy.getUiTranslation("Alert marked as unavailable", "ar"), "تم تحديث حالة التنبيه إلى: غير متوفر");
});

function request(body) {
  return new Request("http://localhost/api/translate", { method: "POST", body: JSON.stringify(body), headers: { "Content-Type": "application/json" } });
}

test("translation API serves known UI copy without the external service", async () => {
  const { POST } = loadTs("app/api/translate/route.ts", { fetch: () => assert.fail("Known copy must work offline") });
  const response = await POST(request({ target: "ar", texts: ["Medications", "Create Alert", "Quantity:"] }));
  assert.equal(response.status, 200);
  assert.deepEqual(await response.json(), { translations: ["الأدوية", "إنشاء تنبيه", "الكمية:"] });
});

test("translation API auto-detects source language for unknown content and keeps batch order", async () => {
  const { POST } = loadTs("app/api/translate/route.ts", {
    fetch: async (_, options) => {
      assert.deepEqual(JSON.parse(options.body), { q: ["Description française"], source: "auto", target: "ar", format: "text" });
      return Response.json({ translatedText: "وصف بالعربية" });
    },
  });
  const response = await POST(request({ target: "ar", texts: ["Medications", "Description française", "Search"] }));
  assert.deepEqual(await response.json(), { translations: ["الأدوية", "وصف بالعربية", "بحث"] });
});

test("translation API rejects invalid input and unusable service results", async () => {
  const { POST } = loadTs("app/api/translate/route.ts", { fetch: async () => Response.json({ translatedText: "" }) });
  assert.equal((await POST(request({ target: "de", texts: ["Search"] }))).status, 400);
  assert.equal((await POST(request({ target: "ar", texts: [] }))).status, 400);
  assert.equal((await POST(request({ target: "ar", texts: [123] }))).status, 400);
  assert.equal((await POST(request({ target: "ar", texts: ["x".repeat(1501)] }))).status, 400);
  assert.equal((await POST(request({ target: "ar", texts: ["Unknown description"] }))).status, 502);
});

// A small DOM fixture exercises the translator's public lifecycle without
// requiring a running app, browser installation, or translation server.
function domFixture() {
  let observer;
  class Element {
    constructor(tag, attributes = {}) { this.tag = tag; this.attributes = new Map(Object.entries(attributes)); this.children = []; this.parentElement = null; this.isConnected = true; }
    append(child) { child.parentElement = this; this.children.push(child); return child; }
    getAttribute(name) { return this.attributes.get(name) ?? null; }
    setAttribute(name, value) { this.attributes.set(name, value); }
    querySelectorAll() { return this.children.flatMap((child) => child instanceof Element ? [child, ...child.querySelectorAll()] : []); }
    closest(selector) {
      const ignored = this.attributes.has("data-translate-ignore") || this.getAttribute("translate") === "no" || this.getAttribute("class") === "notranslate" || this.getAttribute("contenteditable") === "true";
      const matches = selector.split(",").map((s) => s.trim()).includes(this.tag) || (selector.includes("[data-translate-ignore]") && ignored);
      return matches ? this : this.parentElement?.closest(selector) ?? null;
    }
  }
  class Text {
    constructor(value) { this.nodeValue = value; this.isConnected = true; this.parentElement = null; }
  }
  const body = new Element("body");
  const document = {
    body, documentElement: {},
    createTreeWalker(root) {
      const texts = [];
      const visit = (node) => { for (const child of node.children ?? []) { if (child instanceof Text) texts.push(child); else visit(child); } };
      visit(root); let index = 0;
      return { nextNode: () => texts[index++] ?? null };
    },
  };
  const storage = new Map();
  const translator = loadTs("lib/auto-translate-dom.ts", {
    Element, Text, DocumentFragment: class {}, NodeFilter: { SHOW_TEXT: 4 }, document,
    window: { setTimeout, clearTimeout, sessionStorage: { getItem: (key) => storage.get(key) ?? null, setItem: (key, value) => storage.set(key, value) } },
    MutationObserver: class { constructor(callback) { observer = callback; } observe() {} disconnect() {} },
    fetch: () => assert.fail("All fixture content should be translated locally"),
  });
  return { ...translator, body, document, Element, Text, mutate: (mutation) => observer([mutation]) };
}

test("switching English → Arabic → French → English restores original copy and direction", () => {
  const dom = domFixture();
  const text = dom.body.append(new dom.Element("button")).append(new dom.Text("  Create Alert  "));
  let stop = dom.startAutomaticTranslation("ar", () => {});
  assert.equal(text.nodeValue, "  إنشاء تنبيه  ");
  assert.equal(dom.document.documentElement.dir, "rtl");
  stop(); stop = dom.startAutomaticTranslation("fr", () => {});
  assert.equal(text.nodeValue, "  Créer une alerte  ");
  stop(); stop = dom.startAutomaticTranslation("en", () => {});
  assert.equal(text.nodeValue, "  Create Alert  ");
  assert.equal(dom.document.documentElement.dir, "ltr");
  stop();
});

test("textarea placeholders, image alt text, dynamic dialogs and React updates translate; user input stays intact", () => {
  const dom = domFixture();
  const textarea = dom.body.append(new dom.Element("textarea", { placeholder: "Search...", "aria-label": "Search" }));
  const input = textarea.append(new dom.Text("My private draft"));
  const image = dom.body.append(new dom.Element("img", { alt: "Pharmacy illustration" }));
  const ignored = dom.body.append(new dom.Element("div", { translate: "no" })).append(new dom.Text("Search"));
  const stop = dom.startAutomaticTranslation("ar", () => {});
  assert.equal(textarea.getAttribute("placeholder"), "ابحث...");
  assert.equal(textarea.getAttribute("aria-label"), "بحث");
  assert.equal(image.getAttribute("alt"), "رسم توضيحي لصيدلية");
  assert.equal(input.nodeValue, "My private draft");
  assert.equal(ignored.nodeValue, "Search");
  const dialog = dom.body.append(new dom.Element("div"));
  const label = dialog.append(new dom.Text("Confirm & Take"));
  dom.mutate({ type: "childList", addedNodes: [dialog] });
  assert.equal(label.nodeValue, "تأكيد الاستلام");
  label.nodeValue = "Pickup confirmed";
  dom.mutate({ type: "characterData", target: label });
  assert.equal(label.nodeValue, "تم تأكيد الاستلام");
  stop();
});
