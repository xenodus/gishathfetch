import * as SecureStore from "expo-secure-store";

import {
  API_SEARCH_URL,
  API_SESSION_URL,
  MOBILE_CLIENT_HEADER,
} from "../config";
import type { SearchResponse, SessionResponse } from "./types";

const TOKEN_KEY = "gf_api_token";

export class ApiNotReadyError extends Error {
  constructor(message = "Live API search is not available until mobile auth ships on the backend.") {
    super(message);
    this.name = "ApiNotReadyError";
  }
}

function mobileHeaders(token?: string | null): Record<string, string> {
  const headers: Record<string, string> = {
    "X-Client": MOBILE_CLIENT_HEADER,
  };
  if (token) {
    headers.Authorization = `Bearer ${token}`;
  }
  return headers;
}

/**
 * Mint a mobile session token. Requires backend support for X-Client + JSON token response.
 * Until that lands, this will fail against production API (origin verify + cookie-only session).
 */
export async function ensureSession(): Promise<void> {
  const res = await fetch(API_SESSION_URL, {
    method: "GET",
    headers: mobileHeaders(),
  });

  if (!res.ok) {
    throw new Error(`API session failed (${res.status})`);
  }

  const body = (await res.json()) as SessionResponse;
  if (!body.token) {
    throw new ApiNotReadyError("Session response did not include a mobile token.");
  }

  await SecureStore.setItemAsync(TOKEN_KEY, body.token);
}

export async function clearSession(): Promise<void> {
  await SecureStore.deleteItemAsync(TOKEN_KEY);
}

/**
 * Search selected stores for a card name. Throws ApiNotReadyError when the backend
 * has not yet enabled mobile bearer-token auth.
 */
export async function searchCards(
  query: string,
  stores: string[],
): Promise<SearchResponse> {
  const token = await SecureStore.getItemAsync(TOKEN_KEY);
  if (!token) {
    throw new ApiNotReadyError("No mobile session token. Call ensureSession() first.");
  }

  const params = new URLSearchParams({ s: query });
  if (stores.length > 0) {
    params.set("lgs", stores.join(","));
  }

  const res = await fetch(`${API_SEARCH_URL}?${params}`, {
    headers: mobileHeaders(token),
  });

  if (!res.ok) {
    if (res.status === 403) {
      throw new ApiNotReadyError(
        "Search rejected (403). Mobile bearer-token auth is likely not enabled yet.",
      );
    }
    throw new Error(`Search failed (${res.status})`);
  }

  return (await res.json()) as SearchResponse;
}

/** Scryfall autocomplete — works without backend changes. */
export async function fetchAutocompleteSuggestions(query: string): Promise<string[]> {
  const trimmed = query.trim();
  if (trimmed.length < MIN_AUTOCOMPLETE_LENGTH) {
    return [];
  }

  const url = `https://api.scryfall.com/cards/autocomplete?q=${encodeURIComponent(trimmed)}`;
  const res = await fetch(url);
  if (!res.ok) {
    return [];
  }

  const body = (await res.json()) as { data?: string[] };
  return body.data ?? [];
}

const MIN_AUTOCOMPLETE_LENGTH = 2;
