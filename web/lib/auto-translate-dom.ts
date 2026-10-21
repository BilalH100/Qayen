import { getUiTranslation } from "./translation-copy";

export type AppLanguage = "en" | "fr" | "ar";
export type TranslationStatus = "ready" | "error";

type StoredTranslation = {
  original: string;
  translated: string;
  language: AppLanguage;
};

type AttributeState = Map<string, StoredTranslation>;
type PendingTranslation = {
  original: string;
  language: Exclude<AppLanguage, "en">;
  callbacks: Array<(translated: string) => void>;
};

type TranslateApiResponse = {
  translations?: unknown;
  error?: string;
};

const STORAGE_LANGUAGE_KEY = "qayen-language";
const CACHE_PREFIX = "qayen-auto-translation-v2:";
const TRANSLATABLE_ATTRIBUTES = [
  "placeholder",
  "title",
  "aria-label",
  "aria-description",
  "aria-placeholder",
  "alt",
] as const;
const MAX_TEXT_LENGTH = 1_500;
const MAX_BATCH_TEXTS = 35;
const MAX_BATCH_CHARS = 8_000;

// WeakMaps preserve each node's original English text without adding props to React-managed DOM.
const textState = new WeakMap<Text, StoredTranslation>();
const attributeState = new WeakMap<Element, AttributeState>();
const memoryCache = new Map<string, string>();

function cacheId(language: Exclude<AppLanguage, "en">, text: string): string {
  // FNV-1a hash keeps localStorage keys compact. The stored source text is checked on read.
  let hash = 0x811c9dc5;
  for (let index = 0; index < text.length; index += 1) {
    hash ^= text.charCodeAt(index);
    hash = Math.imul(hash, 0x01000193);
  }
  return `${CACHE_PREFIX}${language}:${(hash >>> 0).toString(16)}`;
}

function cacheKey(language: Exclude<AppLanguage, "en">, text: string): string {
  return `${language}\u0000${text}`;
}

function getCachedTranslation(
  language: Exclude<AppLanguage, "en">,
  original: string,
): string | undefined {
  const key = cacheKey(language, original);
  const inMemory = memoryCache.get(key);
  if (inMemory !== undefined) return inMemory;

  try {
    const serialized = window.sessionStorage.getItem(cacheId(language, original));
    if (!serialized) return undefined;
    const parsed = JSON.parse(serialized) as { original?: unknown; translated?: unknown };
    if (parsed.original === original && typeof parsed.translated === "string") {
      memoryCache.set(key, parsed.translated);
      return parsed.translated;
    }
  } catch {
    // Private browsing / storage quota restrictions should not prevent translation.
  }
  return undefined;
}

function saveCachedTranslation(
  language: Exclude<AppLanguage, "en">,
  original: string,
  translated: string,
): void {
  memoryCache.set(cacheKey(language, original), translated);
  try {
    // Session-only caching avoids keeping translated user-specific content across browser sessions.
    window.sessionStorage.setItem(
      cacheId(language, original),
      JSON.stringify({ original, translated }),
    );
  } catch {
    // Keep the in-memory cache if localStorage is unavailable or full.
  }
}

function isIgnoredElement(element: Element | null): boolean {
  if (!element) return true;
  return Boolean(
    element.closest(
      '[data-translate-ignore], [translate="no"], .notranslate, script, style, noscript, code, pre, kbd, samp, svg, [contenteditable="true"]',
    ),
  );
}

function isCandidateText(value: string): boolean {
  const trimmed = value.trim();
  if (!trimmed || trimmed.length > MAX_TEXT_LENGTH) return false;
  if (!/[A-Za-zÀ-ÿ]/.test(trimmed)) return false;
  if (/^(https?:\/\/|www\.)/i.test(trimmed)) return false;
  if (/^[\w.+-]+@[\w.-]+\.[A-Za-z]{2,}$/.test(trimmed)) return false;
  if (/^\+?[\d\s().-]{7,}$/.test(trimmed)) return false;
  return true;
}

function splitWhitespace(value: string): { leading: string; core: string; trailing: string } {
  const leading = value.match(/^\s*/)?.[0] ?? "";
  const trailing = value.match(/\s*$/)?.[0] ?? "";
  return {
    leading,
    trailing,
    core: value.slice(leading.length, value.length - trailing.length),
  };
}

function getOriginalText(node: Text): string {
  const current = node.nodeValue ?? "";
  const saved = textState.get(node);
  if (saved && current === saved.translated) return saved.original;
  textState.set(node, { original: current, translated: current, language: "en" });
  return current;
}

function applyText(node: Text, original: string, translated: string, language: AppLanguage): void {
  if (!node.isConnected) return;
  const current = node.nodeValue ?? "";
  const saved = textState.get(node);
  const currentOriginal = saved && current === saved.translated ? saved.original : current;
  if (currentOriginal !== original) return; // React updated this node while the request was in flight.

  if (current !== translated) node.nodeValue = translated;
  textState.set(node, { original, translated, language });
}

function getOriginalAttribute(element: Element, attribute: string): string | null {
  const current = element.getAttribute(attribute);
  if (current === null) return null;

  const savedAttributes = attributeState.get(element);
  const saved = savedAttributes?.get(attribute);
  if (saved && current === saved.translated) return saved.original;

  const nextAttributes = savedAttributes ?? new Map<string, StoredTranslation>();
  nextAttributes.set(attribute, { original: current, translated: current, language: "en" });
  attributeState.set(element, nextAttributes);
  return current;
}

function applyAttribute(
  element: Element,
  attribute: string,
  original: string,
  translated: string,
  language: AppLanguage,
): void {
  if (!element.isConnected) return;
  const current = element.getAttribute(attribute);
  if (current === null) return;

  const saved = attributeState.get(element)?.get(attribute);
  const currentOriginal = saved && current === saved.translated ? saved.original : current;
  if (currentOriginal !== original) return;

  if (current !== translated) element.setAttribute(attribute, translated);
  const states = attributeState.get(element) ?? new Map<string, StoredTranslation>();
  states.set(attribute, { original, translated, language });
  attributeState.set(element, states);
}

function getElements(root: Node): Element[] {
  const elements: Element[] = [];
  if (root instanceof Element) elements.push(root);
  if (root instanceof Element || root instanceof DocumentFragment) {
    elements.push(...Array.from(root.querySelectorAll("*")));
  }
  return elements;
}

/** Start translating all rendered text and common text-bearing attributes in the document. */
export function startAutomaticTranslation(
  language: AppLanguage,
  onServiceError: (failed: boolean) => void,
): () => void {
  let active = true;
  let scheduled = false;
  let flushing = false;
  let hadRequestError = false;
  let flushTimer: number | undefined;
  const controller = new AbortController();
  const pending = new Map<string, PendingTranslation>();

  const queueTranslation = (
    original: string,
    target: Exclude<AppLanguage, "en">,
    callback: (translated: string) => void,
  ) => {
    const curated = getUiTranslation(original, target);
    if (curated !== undefined) {
      if (active) callback(curated);
      return;
    }
    const cached = getCachedTranslation(target, original);
    if (cached !== undefined) {
      if (active) callback(cached);
      return;
    }

    const key = cacheKey(target, original);
    const existing = pending.get(key);
    if (existing) {
      existing.callbacks.push(callback);
    } else {
      pending.set(key, { original, language: target, callbacks: [callback] });
    }
    scheduleFlush();
  };

  const processTextNode = (node: Text) => {
    if (!active || isIgnoredElement(node.parentElement) || node.parentElement?.closest("textarea")) return;
    const current = node.nodeValue ?? "";
    const saved = textState.get(node);
    if (saved && current === saved.translated && saved.language === language) return;

    const original = getOriginalText(node);
    if (language === "en") {
      applyText(node, original, original, "en");
      return;
    }
    if (!isCandidateText(original)) return;

    const { leading, core, trailing } = splitWhitespace(original);
    if (!core) return;
    queueTranslation(core, language, (translated) => {
      if (active) applyText(node, original, `${leading}${translated}${trailing}`, language);
    });
  };

  const processAttribute = (element: Element, attribute: string) => {
    if (!active || isIgnoredElement(element)) return;
    const current = element.getAttribute(attribute);
    if (current === null) return;
    const saved = attributeState.get(element)?.get(attribute);
    if (saved && current === saved.translated && saved.language === language) return;

    const original = getOriginalAttribute(element, attribute);
    if (original === null) return;
    if (language === "en") {
      applyAttribute(element, attribute, original, original, "en");
      return;
    }
    if (!isCandidateText(original)) return;

    const { leading, core, trailing } = splitWhitespace(original);
    if (!core) return;
    queueTranslation(core, language, (translated) => {
      if (active) applyAttribute(element, attribute, original, `${leading}${translated}${trailing}`, language);
    });
  };

  const processElementAttributes = (element: Element) => {
    if (isIgnoredElement(element)) return;
    for (const attribute of TRANSLATABLE_ATTRIBUTES) processAttribute(element, attribute);
  };

  const scanSubtree = (root: Node) => {
    if (!active) return;
    if (root instanceof Text) {
      processTextNode(root);
      return;
    }

    for (const element of getElements(root)) processElementAttributes(element);

    const walker = document.createTreeWalker(root, NodeFilter.SHOW_TEXT);
    let current: Node | null;
    while ((current = walker.nextNode())) processTextNode(current as Text);
  };

  async function flushQueue(): Promise<void> {
    if (!active || flushing) return;
    flushing = true;
    try {
      while (active && pending.size > 0) {
        const first = pending.values().next().value as PendingTranslation | undefined;
        if (!first) break;
        const batch: PendingTranslation[] = [];
        let totalCharacters = 0;

        for (const [key, item] of pending) {
          if (item.language !== first.language) continue;
          if (batch.length >= MAX_BATCH_TEXTS) break;
          if (batch.length > 0 && totalCharacters + item.original.length > MAX_BATCH_CHARS) break;
          pending.delete(key);
          batch.push(item);
          totalCharacters += item.original.length;
        }
        if (batch.length === 0) {
          // Defensive guard for a future change that introduces mixed batches.
          pending.delete(cacheKey(first.language, first.original));
          continue;
        }

        try {
          const response = await fetch("/api/translate", {
            method: "POST",
            headers: { "Content-Type": "application/json" },
            signal: controller.signal,
            body: JSON.stringify({
              target: first.language,
              texts: batch.map((item) => item.original),
            }),
          });
          if (!response.ok) throw new Error(`Translation request failed (${response.status}).`);
          const payload = (await response.json()) as TranslateApiResponse;
          if (
            !Array.isArray(payload.translations) ||
            payload.translations.length !== batch.length ||
            !payload.translations.every((value) => typeof value === "string" && value.trim().length > 0)
          ) {
            throw new Error("Translation service returned an invalid response.");
          }

          payload.translations.forEach((value, index) => {
            if (typeof value !== "string") return;
            const item = batch[index];
            // An unchanged result may be a proper name or a missed translation.
            // Apply it, but let a later visit retry instead of caching it indefinitely.
            if (value !== item.original) saveCachedTranslation(item.language, item.original, value);
            item.callbacks.forEach((callback) => callback(value));
          });
          hadRequestError = false;
          if (active) onServiceError(false);
        } catch {
          hadRequestError = true;
          if (active) onServiceError(true);
          // Leave the original text in place. Switching languages retries failed text.
        }
      }
    } finally {
      flushing = false;
      if (active && pending.size > 0) scheduleFlush();
      if (active && pending.size === 0 && !hadRequestError) onServiceError(false);
    }
  }

  function scheduleFlush(): void {
    if (scheduled || !active) return;
    scheduled = true;
    flushTimer = window.setTimeout(() => {
      scheduled = false;
      void flushQueue();
    }, 80);
  }

  document.documentElement.lang = language;
  document.documentElement.dir = language === "ar" ? "rtl" : "ltr";
  scanSubtree(document.body);

  const observer = new MutationObserver((mutations) => {
    if (!active) return;
    for (const mutation of mutations) {
      if (mutation.type === "characterData") {
        processTextNode(mutation.target as Text);
      } else if (mutation.type === "attributes") {
        processAttribute(mutation.target as Element, mutation.attributeName ?? "");
      } else {
        mutation.addedNodes.forEach(scanSubtree);
      }
    }
  });

  observer.observe(document.body, {
    childList: true,
    subtree: true,
    characterData: true,
    attributes: true,
    attributeFilter: [...TRANSLATABLE_ATTRIBUTES],
  });

  return () => {
    active = false;
    controller.abort();
    window.clearTimeout(flushTimer);
    observer.disconnect();
    pending.clear();
  };
}

export const LANGUAGE_STORAGE_KEY = STORAGE_LANGUAGE_KEY;
