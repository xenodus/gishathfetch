import { useCallback, useEffect, useRef, useState } from "react";
import {
  loadFavouriteStoresFromStorage,
  persistFavouriteStores,
  storeSelectionsMatch,
  validateFavouriteStores,
} from "../utils/favouriteStores";

const FAVOURITE_STORES_FEEDBACK_DURATION_MS = 2500;

export default function useFavouriteStores() {
  const [favouriteStores, setFavouriteStores] = useState(
    loadFavouriteStoresFromStorage,
  );
  const [favouriteStoresFeedback, setFavouriteStoresFeedback] = useState(null);
  const feedbackTimeoutRef = useRef(null);

  const showFavouriteStoresFeedback = useCallback((message) => {
    if (feedbackTimeoutRef.current) {
      clearTimeout(feedbackTimeoutRef.current);
    }

    setFavouriteStoresFeedback(message);
    feedbackTimeoutRef.current = setTimeout(() => {
      setFavouriteStoresFeedback(null);
      feedbackTimeoutRef.current = null;
    }, FAVOURITE_STORES_FEEDBACK_DURATION_MS);
  }, []);

  useEffect(() => {
    return () => {
      if (feedbackTimeoutRef.current) {
        clearTimeout(feedbackTimeoutRef.current);
      }
    };
  }, []);

  const saveFavourites = useCallback(
    (stores) => {
      const validated = validateFavouriteStores(stores);
      persistFavouriteStores(validated);
      setFavouriteStores(validated);

      if (validated.length === 0) {
        showFavouriteStoresFeedback("Favourite stores cleared");
        return;
      }

      const storeLabel = validated.length === 1 ? "store" : "stores";
      showFavouriteStoresFeedback(
        `${validated.length} ${storeLabel} saved as favourites`,
      );
    },
    [showFavouriteStoresFeedback],
  );

  const favouritesMatchSelection = useCallback(
    (selectedStores) => storeSelectionsMatch(favouriteStores, selectedStores),
    [favouriteStores],
  );

  return {
    favouriteStores,
    hasFavourites: favouriteStores.length > 0,
    saveFavourites,
    favouritesMatchSelection,
    favouriteStoresFeedback,
  };
}
