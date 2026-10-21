import { NextRequest, NextResponse } from "next/server";
import { getUiTranslation, type TranslationLanguage } from "@/lib/translation-copy";

export const runtime = "nodejs";
export const dynamic = "force-dynamic";

const ALLOWED_TARGETS = new Set(["fr", "ar"]);
const MAX_TEXTS = 35;
const MAX_TEXT_LENGTH = 1_500;
const MAX_TOTAL_CHARACTERS = 8_000;

export async function POST(request: NextRequest) {
  let body: unknown;
  try {
    body = await request.json();
  } catch {
    return NextResponse.json({ error: "Invalid JSON request." }, { status: 400 });
  }

  if (!body || typeof body !== "object") {
    return NextResponse.json({ error: "A request body is required." }, { status: 400 });
  }

  const { target, texts } = body as { target?: unknown; texts?: unknown };
  if (typeof target !== "string" || !ALLOWED_TARGETS.has(target)) {
    return NextResponse.json({ error: "Target language must be fr or ar." }, { status: 400 });
  }
  if (
    !Array.isArray(texts) ||
    texts.length === 0 ||
    texts.length > MAX_TEXTS ||
    !texts.every((text) => typeof text === "string" && text.length <= MAX_TEXT_LENGTH)
  ) {
    return NextResponse.json({ error: "The translation batch is invalid or too large." }, { status: 400 });
  }

  const normalizedTexts = texts as string[];
  const totalCharacters = normalizedTexts.reduce((total, text) => total + text.length, 0);
  if (totalCharacters > MAX_TOTAL_CHARACTERS) {
    return NextResponse.json({ error: "The translation batch is too large." }, { status: 413 });
  }

  const translations = normalizedTexts.map((text) => getUiTranslation(text, target as TranslationLanguage));
  const missingIndexes = normalizedTexts.flatMap((_, index) => translations[index] === undefined ? [index] : []);
  if (missingIndexes.length === 0) return NextResponse.json({ translations });

  const serviceBaseUrl = (process.env.LIBRETRANSLATE_URL || "http://127.0.0.1:5000").replace(/\/+$/, "");

  try {
    const upstreamResponse = await fetch(`${serviceBaseUrl}/translate`, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      // Medication descriptions and pharmacy data can originate in French.
      body: JSON.stringify({ q: missingIndexes.map((index) => normalizedTexts[index]), source: "auto", target, format: "text" }),
      signal: AbortSignal.timeout(30_000),
      cache: "no-store",
    });

    if (!upstreamResponse.ok) {
      return NextResponse.json(
        { error: `Translation service returned status ${upstreamResponse.status}.` },
        { status: 502 },
      );
    }

    const result = (await upstreamResponse.json()) as { translatedText?: unknown };
    const rawTranslations = result.translatedText;
    const automaticTranslations = Array.isArray(rawTranslations)
      ? rawTranslations
      : typeof rawTranslations === "string" && missingIndexes.length === 1
        ? [rawTranslations]
        : null;

    if (
      !automaticTranslations ||
      automaticTranslations.length !== missingIndexes.length ||
      !automaticTranslations.every((translation) => typeof translation === "string" && translation.trim().length > 0)
    ) {
      return NextResponse.json({ error: "Translation service returned an invalid response." }, { status: 502 });
    }

    missingIndexes.forEach((index, offset) => { translations[index] = automaticTranslations[offset]; });

    return NextResponse.json({ translations });
  } catch {
    return NextResponse.json(
      { error: "Could not reach LibreTranslate. Start the translation service and try again." },
      { status: 502 },
    );
  }
}
