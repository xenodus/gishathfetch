import { API_BASE_URL } from "../constants";

export const fetchCardKingdomLatestPrice = async (cardName, signal) => {
  const response = await fetch(
    `${API_BASE_URL}ck-price?s=${encodeURIComponent(cardName)}`,
    { signal },
  );
  if (!response.ok) {
    throw new Error("Card Kingdom price lookup is unavailable.");
  }

  const payload = await response.json();
  return payload.data ?? null;
};
