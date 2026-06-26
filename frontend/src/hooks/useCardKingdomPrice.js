import { useEffect, useState } from "react";
import { fetchCardKingdomLatestPrice } from "../utils/cardKingdomPrice";

const useCardKingdomPrice = ({
  searchQuery,
  hasSearched,
  isSearching,
  searchError,
}) => {
  const [cardKingdomPrice, setCardKingdomPrice] = useState(null);
  const [isLoading, setIsLoading] = useState(false);

  useEffect(() => {
    if (!hasSearched || isSearching || searchError || !searchQuery.trim()) {
      setCardKingdomPrice(null);
      setIsLoading(false);
      return;
    }

    const abortController = new AbortController();
    let cancelled = false;

    const loadPrice = async () => {
      setIsLoading(true);
      setCardKingdomPrice(null);

      try {
        const price = await fetchCardKingdomLatestPrice(
          searchQuery,
          abortController.signal,
        );
        if (!cancelled) {
          setCardKingdomPrice(price);
        }
      } catch (error) {
        if (error.name !== "AbortError" && !cancelled) {
          setCardKingdomPrice(null);
        }
      } finally {
        if (!cancelled) {
          setIsLoading(false);
        }
      }
    };

    loadPrice();

    return () => {
      cancelled = true;
      abortController.abort();
    };
  }, [searchQuery, hasSearched, isSearching, searchError]);

  return {
    cardKingdomPrice,
    isCardKingdomLoading: isLoading,
  };
};

export default useCardKingdomPrice;
