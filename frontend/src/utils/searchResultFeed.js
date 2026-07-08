import { cardIdentityKey } from "./cardIdentity.js";

export function searchResultCardKey(card) {
  return `${cardIdentityKey(card)}-${card.url}-${card.price}`;
}

export function computeInFeedAdSlots(
  resultCount,
  adDisplayInterval,
  maxInFeedAds,
) {
  if (resultCount <= adDisplayInterval) {
    return [];
  }

  const slots = [];
  for (let cardIndex = 0; cardIndex < resultCount; cardIndex++) {
    const position = cardIndex + 1;
    if (position % adDisplayInterval !== 0 || position === resultCount) {
      continue;
    }
    if (slots.length >= maxInFeedAds) {
      break;
    }
    slots.push({ cardIndex, slotIndex: slots.length });
  }
  return slots;
}

export function buildSearchResultFeedItems(
  cards,
  { adDisplayInterval, maxInFeedAds },
) {
  const adSlots = computeInFeedAdSlots(
    cards.length,
    adDisplayInterval,
    maxInFeedAds,
  );
  const adSlotIndexByCard = new Map(
    adSlots.map(({ cardIndex, slotIndex }) => [cardIndex, slotIndex]),
  );

  const items = [];
  for (let cardIndex = 0; cardIndex < cards.length; cardIndex++) {
    items.push({ type: "card", card: cards[cardIndex], cardIndex });

    const slotIndex = adSlotIndexByCard.get(cardIndex);
    if (slotIndex !== undefined) {
      items.push({ type: "ad", slotIndex });
    }
  }
  return items;
}
