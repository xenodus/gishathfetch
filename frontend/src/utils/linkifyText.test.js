import assert from "node:assert/strict";
import { splitTextWithLinks } from "./linkifyText.js";

assert.deepEqual(splitTextWithLinks(""), []);

assert.deepEqual(splitTextWithLinks("No links here."), [
  { type: "text", value: "No links here.", start: 0 },
]);

assert.deepEqual(
  splitTextWithLinks(
    "Gishath Fetch is on Telegram. Open t.me/GishathFetchBot and try /price Lightning Bolt.",
  ),
  [
    { type: "text", value: "Gishath Fetch is on Telegram. Open ", start: 0 },
    {
      type: "link",
      value: "t.me/GishathFetchBot",
      href: "https://t.me/GishathFetchBot",
      start: 35,
    },
    { type: "text", value: " and try /price Lightning Bolt.", start: 55 },
  ],
);

assert.deepEqual(splitTextWithLinks("Visit https://gishathfetch.com today."), [
  { type: "text", value: "Visit ", start: 0 },
  {
    type: "link",
    value: "https://gishathfetch.com",
    href: "https://gishathfetch.com/",
    start: 6,
  },
  { type: "text", value: " today.", start: 30 },
]);

assert.deepEqual(splitTextWithLinks("See https://example.com."), [
  { type: "text", value: "See ", start: 0 },
  {
    type: "link",
    value: "https://example.com",
    href: "https://example.com/",
    start: 4,
  },
  { type: "text", value: ".", start: 23 },
]);

console.log("linkifyText.test.js: all tests passed");
