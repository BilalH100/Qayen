const assert = require("node:assert/strict");
const { test } = require("node:test");
const fs = require("node:fs");
const path = require("node:path");
const vm = require("node:vm");
const ts = require("typescript");

function openProfile() {
  const state = [];
  let index = 0;
  let failed = false;
  const calls = [];
  const user = { id: 7, name: "Original Name", email: "person@example.com", phone: "+212600000000" };
  const auth = { user, updateProfile: async data => {
    calls.push(data);
    if (failed) throw new Error("Could not save your profile. Please try again.");
    Object.assign(user, data);
  } };
  const source = fs.readFileSync(path.join(__dirname, "../app/profile/page.tsx"), "utf8");
  const javascript = ts.transpileModule(source, { compilerOptions: { module: ts.ModuleKind.CommonJS, target: ts.ScriptTarget.ES2020, jsx: ts.JsxEmit.ReactJSX } }).outputText;
  const value = { exports: {} };
  const jsx = (type, props) => ({ type, props });
  vm.runInNewContext(javascript, { module: value, exports: value.exports, require(name) {
    if (name === "react/jsx-runtime") return { jsx, jsxs: jsx };
    if (name === "react") return { useState(initial) {
      const slot = index++;
      if (!(slot in state)) state[slot] = initial;
      return [state[slot], next => { state[slot] = typeof next === "function" ? next(state[slot]) : next; }];
    } };
    if (name === "@/contexts/auth-context") return { useAuth: () => auth };
    return new Proxy({}, { get: (_, property) => String(property) });
  } });
  const Content = value.exports.default().props.children.type;
  const render = () => { index = 0; return Content(); };
  const find = (tree, predicate) => {
    if (!tree || typeof tree !== "object") return undefined;
    if (predicate(tree)) return tree;
    const children = tree.props?.children;
    for (const child of Array.isArray(children) ? children : [children]) {
      const found = find(child, predicate);
      if (found) return found;
    }
  };
  return { render, find, calls, user, fail: () => { failed = true; } };
}

test("Edit Profile opens personal details, saves changes and closes after success", async () => {
  const page = openProfile();
  let tree = page.render();
  page.find(tree, node => node.type === "Button" && node.props.onClick && Array.isArray(node.props.children)).props.onClick();
  tree = page.render();
  assert.equal(page.find(tree, node => node.type === "Dialog").props.open, true);
  assert.equal(page.find(tree, node => node.props?.id === "profile-name").props.value, "Original Name");
  assert.equal(page.find(tree, node => node.props?.id === "profile-email").props.readOnly, true);
  page.find(tree, node => node.props?.id === "profile-name").props.onChange({ target: { value: " New Name " } });
  tree = page.render();
  await page.find(tree, node => node.type === "form").props.onSubmit({ preventDefault() {} });
  assert.equal(page.calls.length, 1);
  assert.equal(page.calls[0].name, "New Name");
  assert.equal(page.calls[0].phone, "+212600000000");
  assert.equal(page.find(page.render(), node => node.type === "Dialog").props.open, false);
});

test("Cancel does not save and a failed save keeps the editor open with feedback", async () => {
  const page = openProfile();
  const open = () => page.find(page.render(), node => node.type === "Button" && node.props.onClick && Array.isArray(node.props.children)).props.onClick();
  open();
  page.find(page.render(), node => node.type === "Button" && node.props.children === "Cancel").props.onClick();
  assert.equal(page.calls.length, 0);
  assert.equal(page.find(page.render(), node => node.type === "Dialog").props.open, false);
  open();
  page.fail();
  await page.find(page.render(), node => node.type === "form").props.onSubmit({ preventDefault() {} });
  const tree = page.render();
  assert.equal(page.find(tree, node => node.type === "Dialog").props.open, true);
  assert.ok(page.find(tree, node => node.type === "Alert" && node.props.role === "alert"));
});
