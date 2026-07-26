import { dedupeCartItems, normalizeExtraInfo } from "./cardIdentity.js";

export const CART_TRANSFER_PREFIX = "gish-cart-v1:";

const normalizeText = (value) =>
  String(value ?? "")
    .trim()
    .normalize("NFKC");

/**
 * @param {unknown} raw
 * @returns {object | null}
 */
export function sanitizeCartItem(raw) {
  if (!raw || typeof raw !== "object" || Array.isArray(raw)) {
    return null;
  }

  const name = normalizeText(raw.name);
  const src = normalizeText(raw.src);
  const url = typeof raw.url === "string" ? raw.url.trim() : "";
  if (!name || !src || !url) {
    return null;
  }

  const price = Number(raw.price);
  if (!Number.isFinite(price)) {
    return null;
  }

  const item = {
    name,
    src,
    url,
    price,
    inStock: raw.inStock !== false,
    isFoil: !!raw.isFoil,
  };

  if (typeof raw.img === "string" && raw.img.trim()) {
    item.img = raw.img.trim();
  }
  if (typeof raw.quality === "string" && raw.quality.trim()) {
    item.quality = raw.quality.trim();
  }
  const extraInfo = normalizeExtraInfo(raw.extraInfo);
  if (extraInfo) {
    item.extraInfo = extraInfo;
  }
  const savedAt = Number(raw.savedAt);
  if (Number.isFinite(savedAt) && savedAt > 0) {
    item.savedAt = savedAt;
  }

  return item;
}

/**
 * @param {unknown} items
 * @returns {object[]}
 */
export function validateCartItems(items) {
  if (!Array.isArray(items)) {
    return [];
  }

  const validated = [];
  for (const entry of items) {
    const item = sanitizeCartItem(entry);
    if (item) {
      validated.push(item);
    }
  }
  return validated;
}

/**
 * @param {object[]} cart
 * @returns {string}
 */
export function encodeCartExport(cart) {
  const payload = validateCartItems(cart);
  const json = JSON.stringify(payload);
  const bytes = new TextEncoder().encode(json);
  let binary = "";
  for (const byte of bytes) {
    binary += String.fromCharCode(byte);
  }
  const base64 = btoa(binary)
    .replace(/\+/g, "-")
    .replace(/\//g, "_")
    .replace(/=+$/g, "");
  return `${CART_TRANSFER_PREFIX}${base64}`;
}

/**
 * @param {string} encoded
 * @returns {{ items: object[], error: string | null }}
 */
export function decodeCartImport(encoded) {
  const trimmed = String(encoded ?? "").trim();
  if (!trimmed) {
    return { items: [], error: "Paste an export code to import." };
  }

  if (!trimmed.startsWith(CART_TRANSFER_PREFIX)) {
    return { items: [], error: "Unrecognized export code." };
  }

  const base64Url = trimmed.slice(CART_TRANSFER_PREFIX.length);
  if (!base64Url) {
    return { items: [], error: "Export code is empty." };
  }

  try {
    const base64 = base64Url.replace(/-/g, "+").replace(/_/g, "/");
    const padding =
      base64.length % 4 === 0 ? "" : "=".repeat(4 - (base64.length % 4));
    const binary = atob(base64 + padding);
    const bytes = Uint8Array.from(binary, (char) => char.charCodeAt(0));
    const json = new TextDecoder().decode(bytes);
    const parsed = JSON.parse(json);
    const items = validateCartItems(parsed);
    if (items.length === 0) {
      return {
        items: [],
        error: "No valid saved cards found in this export code.",
      };
    }
    return { items, error: null };
  } catch {
    return { items: [], error: "Could not read this export code." };
  }
}

/**
 * Appends imported items to the current cart and dedupes by card identity.
 *
 * @param {object[]} current
 * @param {object[]} imported
 * @returns {object[]}
 */
export function mergeCartImport(current, imported) {
  const validated = validateCartItems(imported);
  if (validated.length === 0) {
    return validateCartItems(current);
  }
  return dedupeCartItems([...validateCartItems(current), ...validated]);
}
