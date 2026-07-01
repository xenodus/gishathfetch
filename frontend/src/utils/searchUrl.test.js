import assert from "node:assert/strict";
import {
  buildSearchUrl,
  buildSearchUrlParams,
  isTrackingParam,
  mergeTrackingParams,
} from "./searchUrl.js";

assert.equal(isTrackingParam("utm_source"), true);
assert.equal(isTrackingParam("UTM_CAMPAIGN"), true);
assert.equal(isTrackingParam("s"), false);
assert.equal(isTrackingParam("lgs"), false);
assert.equal(isTrackingParam("faq"), false);

const existing = new URLSearchParams(
  "s=Old&utm_source=popular_searches&utm_medium=internal&utm_campaign=popular_searches&utm_content=last24Hours&faq=1",
);
const params = buildSearchUrlParams("Loki, Lord of Misrule", [], existing);

assert.equal(params.get("s"), "Loki, Lord of Misrule");
assert.equal(params.get("utm_source"), "popular_searches");
assert.equal(params.get("utm_medium"), "internal");
assert.equal(params.get("utm_campaign"), "popular_searches");
assert.equal(params.get("utm_content"), "last24Hours");
assert.equal(params.has("faq"), false);

const withStores = buildSearchUrlParams(
  "Lightning Bolt",
  ["Hideout"],
  existing,
);
assert.equal(withStores.get("lgs"), "Hideout");
assert.equal(withStores.get("utm_source"), "popular_searches");

const target = new URLSearchParams("s=test&utm_campaign=override");
mergeTrackingParams(
  new URLSearchParams("utm_source=email&utm_campaign=newsletter"),
  target,
);
assert.equal(target.get("utm_source"), "email");
assert.equal(target.get("utm_campaign"), "override");

const url = buildSearchUrl(
  "https://gishathfetch.com/",
  "Loki, Lord of Misrule",
  [],
  existing,
);
assert.ok(url.includes("utm_source=popular_searches"));
assert.ok(url.includes("utm_content=last24Hours"));
assert.ok(url.includes("s=Loki%2C+Lord+of+Misrule"));

console.log("searchUrl tests passed");
