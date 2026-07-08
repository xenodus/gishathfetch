import assert from "node:assert/strict";
import {
  buildSearchResultFeedItems,
  computeInFeedAdSlots,
} from "./searchResultFeed.js";

const feedAdIndices = (feedItems) =>
  feedItems
    .map((item, index) => (item.type === "ad" ? index : -1))
    .filter((index) => index >= 0);

assert.deepEqual(computeInFeedAdSlots(32, 8, 3), [
  { cardIndex: 7, slotIndex: 0 },
  { cardIndex: 15, slotIndex: 1 },
  { cardIndex: 23, slotIndex: 2 },
]);

assert.deepEqual(computeInFeedAdSlots(24, 8, 3), [
  { cardIndex: 7, slotIndex: 0 },
  { cardIndex: 15, slotIndex: 1 },
]);

assert.deepEqual(computeInFeedAdSlots(8, 8, 3), []);
assert.deepEqual(computeInFeedAdSlots(9, 8, 3), [
  { cardIndex: 7, slotIndex: 0 },
]);

const cards = Array.from({ length: 20 }, (_, id) => ({ id }));
const feedOptions = { adDisplayInterval: 8, maxInFeedAds: 3 };
const ascendingFeed = buildSearchResultFeedItems(cards, feedOptions);
const descendingFeed = buildSearchResultFeedItems(
  [...cards].reverse(),
  feedOptions,
);

assert.deepEqual(feedAdIndices(ascendingFeed), feedAdIndices(descendingFeed));
assert.deepEqual(feedAdIndices(ascendingFeed), [8, 17]);

const shuffledCards = [
  cards[3],
  cards[0],
  cards[19],
  ...cards.slice(4, 19),
  cards[1],
  cards[2],
];
const shuffledFeed = buildSearchResultFeedItems(shuffledCards, feedOptions);
assert.deepEqual(feedAdIndices(shuffledFeed), feedAdIndices(ascendingFeed));

console.log("searchResultFeed tests passed");
