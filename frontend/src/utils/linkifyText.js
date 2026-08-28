const URL_PATTERN =
  /\b((?:https?:\/\/)?(?:[a-zA-Z0-9-]+\.)+[a-zA-Z]{2,}(?:\/[^\s<>,;:!?)]*)?)/gi;

const TRAILING_PUNCTUATION = /[.,;:!?)]$/;

function trimTrailingPunctuation(value) {
  let url = value;
  let trailing = "";

  while (url.length > 0 && TRAILING_PUNCTUATION.test(url)) {
    trailing = url.slice(-1) + trailing;
    url = url.slice(0, -1);
  }

  return { url, trailing };
}

function toHref(url) {
  const href = /^https?:\/\//i.test(url) ? url : `https://${url}`;
  try {
    const parsed = new URL(href);
    if (parsed.protocol !== "http:" && parsed.protocol !== "https:") {
      return null;
    }
    return parsed.toString();
  } catch {
    return null;
  }
}

/**
 * Split plain text into alternating text and link segments for safe rendering.
 * Bare domains such as t.me/GishathFetchBot are linked with https://.
 */
export function splitTextWithLinks(text) {
  if (typeof text !== "string" || text === "") {
    return [];
  }

  const parts = [];
  let lastIndex = 0;

  for (const match of text.matchAll(URL_PATTERN)) {
    const rawMatch = match[1];
    const matchIndex = match.index ?? 0;
    const { url, trailing } = trimTrailingPunctuation(rawMatch);
    const href = toHref(url);

    if (!href) {
      continue;
    }

    if (matchIndex > lastIndex) {
      parts.push({
        type: "text",
        value: text.slice(lastIndex, matchIndex),
        start: lastIndex,
      });
    }

    parts.push({ type: "link", value: url, href, start: matchIndex });
    lastIndex = matchIndex + rawMatch.length;

    if (trailing) {
      parts.push({
        type: "text",
        value: trailing,
        start: lastIndex - trailing.length,
      });
    }
  }

  if (lastIndex < text.length) {
    parts.push({
      type: "text",
      value: text.slice(lastIndex),
      start: lastIndex,
    });
  }

  return parts.length > 0 ? parts : [{ type: "text", value: text, start: 0 }];
}
