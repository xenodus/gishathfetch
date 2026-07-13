import { useCallback, useRef, useState } from "react";
import { CK_PRICE_CHANGES_DISPLAY_LIMIT, CK_PRICE_CHANGES_URL } from "../constants";
import { parseCKPriceIncreases } from "../utils/ckPriceChanges";

export default function useCKPriceIncreases() {
  const [priceIncreases, setPriceIncreases] = useState([]);
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState(null);
  const [hasLoaded, setHasLoaded] = useState(false);
  const requestIdRef = useRef(0);

  const loadPriceIncreases = useCallback(async () => {
    const requestId = requestIdRef.current + 1;
    requestIdRef.current = requestId;
    setIsLoading(true);
    setError(null);

    try {
      const response = await fetch(CK_PRICE_CHANGES_URL);
      if (!response.ok) {
        throw new Error(`Failed to load CK price changes (${response.status})`);
      }

      const payload = await response.json();
      if (requestId !== requestIdRef.current) {
        return;
      }

      setPriceIncreases(
        parseCKPriceIncreases(payload, CK_PRICE_CHANGES_DISPLAY_LIMIT),
      );
      setHasLoaded(true);
    } catch (loadError) {
      if (requestId !== requestIdRef.current) {
        return;
      }

      console.error("Failed to load CK price increases:", loadError);
      setPriceIncreases([]);
      setError("Could not load CK price increases.");
      setHasLoaded(true);
    } finally {
      if (requestId === requestIdRef.current) {
        setIsLoading(false);
      }
    }
  }, []);

  return {
    priceIncreases,
    isLoading,
    error,
    hasLoaded,
    loadPriceIncreases,
  };
}
