const SCRYFALL_AUTOCOMPLETE_URL = "https://api.scryfall.com/cards/autocomplete";
const SCRYFALL_NAMED_URL = "https://api.scryfall.com/cards/named";

const canonicalNameCache = new Map();

export function foldCardNameForMatch(name) {
  return String(name ?? "")
    .normalize("NFKD")
    .replace(/[\u0300-\u036f]/g, "")
    .toLowerCase();
}

export function cardNamesMatchIgnoringDiacritics(a, b) {
  const left = String(a ?? "").trim();
  const right = String(b ?? "").trim();
  if (!left || !right) {
    return false;
  }
  if (left.toLowerCase() === right.toLowerCase()) {
    return true;
  }
  return foldCardNameForMatch(left) === foldCardNameForMatch(right);
}

function pickCanonicalCandidate(query, candidates) {
  if (!Array.isArray(candidates) || candidates.length === 0) {
    return null;
  }

  const trimmed = String(query ?? "").trim();
  for (const candidate of candidates) {
    if (candidate === trimmed) {
      return candidate;
    }
  }
  for (const candidate of candidates) {
    if (cardNamesMatchIgnoringDiacritics(candidate, trimmed)) {
      return candidate;
    }
  }
  return null;
}

async function fetchNamedCard(query, matchMode) {
  const url = new URL(SCRYFALL_NAMED_URL);
  url.searchParams.set(matchMode, query);

  const res = await fetch(url);
  if (!res.ok) {
    return null;
  }

  const card = await res.json();
  if (!card?.name || !cardNamesMatchIgnoringDiacritics(card.name, query)) {
    return null;
  }
  return card.name;
}

export async function resolveCanonicalCardName(name) {
  const trimmed = String(name ?? "").trim();
  if (!trimmed) {
    return trimmed;
  }

  const cached = canonicalNameCache.get(trimmed);
  if (cached) {
    return cached;
  }

  let resolved = trimmed;

  try {
    const autocompleteUrl = new URL(SCRYFALL_AUTOCOMPLETE_URL);
    autocompleteUrl.searchParams.set("q", trimmed);
    const autocompleteRes = await fetch(autocompleteUrl);
    if (autocompleteRes.ok) {
      const autocomplete = await autocompleteRes.json();
      const candidate = pickCanonicalCandidate(trimmed, autocomplete?.data);
      if (candidate) {
        resolved = candidate;
      }
    }

    if (resolved === trimmed) {
      const exactName = await fetchNamedCard(trimmed, "exact");
      if (exactName) {
        resolved = exactName;
      }
    }

    if (resolved === trimmed) {
      const fuzzyName = await fetchNamedCard(trimmed, "fuzzy");
      if (fuzzyName) {
        resolved = fuzzyName;
      }
    }
  } catch (err) {
    console.error("Failed to resolve canonical card name:", err);
  }

  canonicalNameCache.set(trimmed, resolved);
  if (resolved !== trimmed) {
    canonicalNameCache.set(resolved, resolved);
  }

  return resolved;
}

export async function resolveCanonicalCardNames(names) {
  const uniqueNames = [
    ...new Set(names.map((name) => String(name ?? "").trim()).filter(Boolean)),
  ];
  const resolved = new Map();

  await Promise.all(
    uniqueNames.map(async (name) => {
      resolved.set(name, await resolveCanonicalCardName(name));
    }),
  );

  return resolved;
}

export function applyCanonicalCardNames(cards, resolvedNames) {
  if (!Array.isArray(cards) || cards.length === 0) {
    return cards;
  }

  let changed = false;
  const updated = cards.map((card) => {
    const canonicalName = resolvedNames.get(card.name);
    if (!canonicalName || canonicalName === card.name) {
      return card;
    }
    changed = true;
    return { ...card, name: canonicalName };
  });

  return changed ? updated : cards;
}
