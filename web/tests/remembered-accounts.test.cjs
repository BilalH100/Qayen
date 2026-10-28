const assert = require("node:assert/strict");
const fs = require("node:fs");
const path = require("node:path");
const vm = require("node:vm");
const { test } = require("node:test");
const ts = require("typescript");

const source = ts.transpileModule(fs.readFileSync(path.join(__dirname, "../lib/remembered-accounts.ts"), "utf8"), {
  compilerOptions: { module: ts.ModuleKind.CommonJS, target: ts.ScriptTarget.ES2020 },
}).outputText;

function openBrowser(storage = new Map(), blocked = false) {
  const module = { exports: {} };
  const events = [];
  const window = {
    localStorage: {
      getItem(key) { if (blocked) throw new Error("Storage unavailable"); return storage.get(key) ?? null; },
      setItem(key, value) { if (blocked) throw new Error("Storage unavailable"); storage.set(key, value); },
    },
    dispatchEvent(event) { events.push(event.type); },
  };
  vm.runInNewContext(source, { module, exports: module.exports, window, Event });
  return { ...module.exports, storage, events };
}

test("all successful account emails survive a new browser session, including more than five accounts", () => {
  const first = openBrowser();
  for (let n = 0; n < 12; n++) assert.equal(first.rememberAccountEmail(`tester${n}@example.com`), true);
  const reopened = openBrowser(first.storage);
  assert.equal(reopened.readRememberedAccounts().length, 12);
  assert.equal(reopened.readRememberedAccounts()[0], "tester11@example.com");
  assert.equal(reopened.readRememberedAccounts()[11], "tester0@example.com");
  assert.equal(first.storage.size, 1);
  assert.ok(JSON.parse(first.storage.get(first.SAVED_ACCOUNTS_KEY)).every(value => typeof value === "string"));
});

test("legacy saved emails are retained, cleaned and moved to the front without duplicates", () => {
  const storage = new Map([["kayena_saved_accounts", JSON.stringify([" TEST@EXAMPLE.COM ", null, 123, {}, "test@example.com", "other@example.com"])]]);
  const browser = openBrowser(storage);
  assert.equal(JSON.stringify(browser.readRememberedAccounts()), JSON.stringify(["test@example.com", "other@example.com"]));
  browser.rememberAccountEmail(" Other@Example.com ");
  assert.equal(JSON.stringify(browser.readRememberedAccounts()), JSON.stringify(["other@example.com", "test@example.com"]));
  assert.equal(browser.events.at(-1), browser.SAVED_ACCOUNTS_CHANGED);
});

test("a malformed saved value can recover on the next successful account save", () => {
  const browser = openBrowser(new Map([["kayena_saved_accounts", "{damaged json"]]));
  assert.equal(browser.readRememberedAccounts().length, 0);
  assert.equal(browser.rememberAccountEmail("valid@example.com"), true);
  assert.equal(openBrowser(browser.storage).readRememberedAccounts()[0], "valid@example.com");
});

test("removing an account persists and storage failures are reported without throwing", () => {
  const browser = openBrowser();
  browser.rememberAccountEmail("first@example.com");
  browser.rememberAccountEmail("second@example.com");
  assert.equal(browser.forgetAccountEmail("FIRST@example.com"), true);
  assert.equal(JSON.stringify(openBrowser(browser.storage).readRememberedAccounts()), JSON.stringify(["second@example.com"]));
  const blocked = openBrowser(new Map(), true);
  assert.equal(blocked.readRememberedAccounts().length, 0);
  assert.equal(blocked.rememberAccountEmail("test@example.com"), false);
  assert.equal(blocked.forgetAccountEmail("test@example.com"), false);
});
