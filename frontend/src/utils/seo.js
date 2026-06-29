import {
  BASE_URL,
  LGS_OPTIONS,
  PAGE_TITLE,
  SITE_DESCRIPTION,
} from "../constants";

function getOrCreateMeta(attrName, attrValue) {
  let element = document.head.querySelector(`meta[${attrName}="${attrValue}"]`);
  if (!element) {
    element = document.createElement("meta");
    element.setAttribute(attrName, attrValue);
    document.head.appendChild(element);
  }
  return element;
}

function setMetaContent(attrName, attrValue, content) {
  getOrCreateMeta(attrName, attrValue).setAttribute("content", content);
}

function getOrCreateLink(rel) {
  let element = document.head.querySelector(`link[rel="${rel}"]`);
  if (!element) {
    element = document.createElement("link");
    element.setAttribute("rel", rel);
    document.head.appendChild(element);
  }
  return element;
}

function setCanonicalUrl(url) {
  getOrCreateLink("canonical").setAttribute("href", url);
}

export function buildSearchPageTitle(query) {
  return `${query} @ Gishath Fetch`;
}

export function buildSearchDescription(query) {
  const storeCount = LGS_OPTIONS.length;
  return `Compare ${query} prices across ${storeCount} Singapore MTG stores and online shops. In-stock results sorted by price.`;
}

export function buildSearchCanonicalUrl(query) {
  const params = new URLSearchParams();
  params.set("s", query);
  return `${BASE_URL}?${params.toString()}`;
}

export function applySearchSeo(query) {
  const title = buildSearchPageTitle(query);
  const description = buildSearchDescription(query);
  const url = buildSearchCanonicalUrl(query);

  document.title = title;
  setMetaContent("name", "description", description);
  setMetaContent("property", "og:title", title);
  setMetaContent("property", "og:description", description);
  setMetaContent("property", "og:url", url);
  setMetaContent("name", "twitter:title", title);
  setMetaContent("name", "twitter:description", description);
  setCanonicalUrl(url);
}

export function applyHomeSeo() {
  document.title = PAGE_TITLE;
  setMetaContent("name", "description", SITE_DESCRIPTION);
  setMetaContent("property", "og:title", PAGE_TITLE);
  setMetaContent("property", "og:description", SITE_DESCRIPTION);
  setMetaContent("property", "og:url", BASE_URL);
  setMetaContent("name", "twitter:title", PAGE_TITLE);
  setMetaContent("name", "twitter:description", SITE_DESCRIPTION);
  setCanonicalUrl(BASE_URL);
}

const MAP_PAGE_TITLE = "Singapore MTG Store Map | Gishath Fetch";
const MAP_PAGE_DESCRIPTION =
  "Find Singapore Magic: The Gathering local game stores on an interactive map. Addresses, websites, and directions for every pinned shop.";

export function applyMapSeo() {
  const mapUrl = `${BASE_URL}map`;

  document.title = MAP_PAGE_TITLE;
  setMetaContent("name", "description", MAP_PAGE_DESCRIPTION);
  setMetaContent("property", "og:title", MAP_PAGE_TITLE);
  setMetaContent("property", "og:description", MAP_PAGE_DESCRIPTION);
  setMetaContent("property", "og:url", mapUrl);
  setMetaContent("name", "twitter:title", MAP_PAGE_TITLE);
  setMetaContent("name", "twitter:description", MAP_PAGE_DESCRIPTION);
  setCanonicalUrl(mapUrl);
}
