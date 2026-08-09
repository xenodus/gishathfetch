import Constants from "expo-constants";

type AppExtra = {
  apiBaseUrl?: string;
  siteBaseUrl?: string;
};

const extra = (Constants.expoConfig?.extra ?? {}) as AppExtra;

function trimTrailingSlash(url: string): string {
  return url.replace(/\/$/, "");
}

export const API_BASE_URL = trimTrailingSlash(
  extra.apiBaseUrl ?? "https://api.gishathfetch.com",
);
export const SITE_BASE_URL = trimTrailingSlash(
  extra.siteBaseUrl ?? "https://gishathfetch.com",
);

export const API_SEARCH_URL = `${API_BASE_URL}/search`;
export const API_SESSION_URL = `${API_BASE_URL}/session`;
export const TOP_SEARCH_KEYWORDS_URL = `${SITE_BASE_URL}/analytics/top-search-keywords/latest.json`;
export const CK_PRICE_CHANGES_URL = `${SITE_BASE_URL}/analytics/ck-price-changes/latest.json`;

/** Sent on API requests until mobile bearer-token auth ships on the backend. */
export const MOBILE_CLIENT_HEADER = "gishathfetch-mobile";
